// Package melodix provides audio playback management for a Discord bot.
package melodix

import (
	"app/internal/config"
	"app/pkg/dca"
	"io"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// Status represents the playback status of the Player.
type Status int32

const (
	StatusResting Status = iota
	StatusPlaying
	StatusPaused
	StatusError
)

// Player manages audio playback and song queue.
type Player struct {
	sync.Mutex
	VoiceConnection  *discordgo.VoiceConnection
	StreamingSession *dca.StreamingSession
	SongQueue        []*Song
	CurrentSong      *Song
	CurrentStatus    Status
	SkipInterrupt    chan bool
}

// IPlayer defines the interface for managing audio playback and song queue.
type IPlayer interface {
	Play(startAt int, song *Song)
	Skip()
	Enqueue(song *Song)
	Dequeue() *Song
	ClearQueue()
	Stop()
	Pause()
	Unpause()
	GetCurrentStatus() Status
	SetCurrentStatus(status Status)
	GetSongQueue() []*Song
	GetVoiceConnection() *discordgo.VoiceConnection
	SetVoiceConnection(voiceConnection *discordgo.VoiceConnection)
	GetStreamingSession() *dca.StreamingSession
	GetCurrentSong() *Song
}

// NewPlayer creates a new Player instance.
func NewPlayer(guildID string) IPlayer {
	return &Player{
		SongQueue:     make([]*Song, 0),
		SkipInterrupt: make(chan bool, 1),
		CurrentStatus: StatusResting,
	}
}

// Skip skips to the next song in the queue.
func (p *Player) Skip() {
	slog.Info("Skipping to next song")

	if p.VoiceConnection == nil {
		return
	}

	p.CurrentStatus = StatusResting

	if len(p.SkipInterrupt) == 0 {
		history := NewHistory()
		history.AddPlaybackCountStats(p.VoiceConnection.GuildID, p.CurrentSong.ID)

		p.SkipInterrupt <- true
		p.Play(0, nil)
	}
}

// Enqueue adds a song to the queue.
func (p *Player) Enqueue(song *Song) {
	slog.Infof("Enqueuing song to queue: %v", song.Name)

	p.Lock()
	defer p.Unlock()

	p.SongQueue = append(p.SongQueue, song)
}

// Dequeue removes and returns the first song from the queue.
func (p *Player) Dequeue() *Song {
	slog.Info("Dequeuing song and returning it from queue")

	p.Lock()
	defer p.Unlock()

	if len(p.SongQueue) == 0 {
		return nil
	}

	firstSong := p.SongQueue[0]
	p.SongQueue = p.SongQueue[1:]

	return firstSong
}

// ClearQueue clears the song queue.
func (p *Player) ClearQueue() {
	slog.Info("Clearing song queue")
	p.SongQueue = make([]*Song, 0)
}

// Stop stops audio playback and disconnects from the voice channel.
func (p *Player) Stop() {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	p.CurrentStatus = StatusResting

	if p.VoiceConnection == nil {
		return
	}

	err := p.VoiceConnection.Disconnect()
	if err != nil {
		slog.Errorf("Error disconnecting voice connection: %v", err)
	}

	p.VoiceConnection = nil
	p.StreamingSession = nil
	p.CurrentSong = nil
}

// Pause pauses audio playback.
func (p *Player) Pause() {
	slog.Info("Pausing audio playback")

	if p.VoiceConnection == nil {
		return
	}

	if p.StreamingSession == nil {
		return
	}

	if p.CurrentStatus == StatusPlaying {
		p.StreamingSession.SetPaused(true)
		p.CurrentStatus = StatusPaused
	}
}

// Unpause resumes audio playback.
func (p *Player) Unpause() {
	slog.Info("Resuming playback")

	if p.StreamingSession != nil {
		if p.CurrentStatus != StatusPlaying {
			p.StreamingSession.SetPaused(false)
			p.CurrentStatus = StatusPlaying
		}
	} else {
		p.Play(0, nil)
		p.CurrentStatus = StatusPlaying
	}
}

// Play starts playing the current or specified song.
func (p *Player) Play(startAt int, song *Song) {

	if song == nil {
		p.CurrentSong = p.Dequeue()
		if p.CurrentSong == nil {
			slog.Info("No songs in queue")
			p.CurrentStatus = StatusResting
			return
		}
	}

	slog.Infof("Playing song: %v", p.CurrentSong.Name)
	slog.Infof("Playing song at: %v", time.Duration(startAt)*time.Second)

	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	options := &dca.EncodeOptions{
		Volume:                  1.0,
		FrameDuration:           config.DcaFrameDuration,
		Bitrate:                 config.DcaBitrate,
		PacketLoss:              config.DcaPacketLoss,
		RawOutput:               config.DcaRawOutput,
		Application:             config.DcaApplication,
		CompressionLevel:        config.DcaCompressionLevel,
		BufferedFrames:          config.DcaBufferedFrames,
		VBR:                     config.DcaVBR,
		StartTime:               startAt,
		ReconnectAtEOF:          config.DcaReconnectAtEOF,
		ReconnectStreamed:       config.DcaReconnectStreamed,
		ReconnectOnNetworkError: config.DcaReconnectOnNetworkError,
		ReconnectOnHttpError:    config.DcaReconnectOnHttpError,
		ReconnectDelayMax:       config.DcaReconnectDelayMax,
		FfmpegBinaryPath:        config.DcaFfmpegBinaryPath,
		EncodingLineLog:         config.DcaEncodingLineLog,
		UserAgent:               config.DcaUserAgent,
	}

	encodingSession, err := dca.EncodeFile(p.CurrentSong.DownloadURL, options)
	if err != nil {
		slog.Errorf("Error encoding song: %v", err)
		return
	}
	defer encodingSession.Cleanup()

	for p.VoiceConnection == nil || !p.VoiceConnection.Ready {
		time.Sleep(100 * time.Millisecond)
	}

	err = p.VoiceConnection.Speaking(true)
	if err != nil {
		slog.Errorf("Error connecting to Discord voice: %v", err)
		p.VoiceConnection.Speaking(false)
		return
	}

	done := make(chan error)
	p.StreamingSession = dca.NewStream(encodingSession, p.VoiceConnection, done)

	slog.Info("Stream is created, waiting for finish or error")

	p.CurrentStatus = StatusPlaying

	history := NewHistory()
	history.AddTrackToHistory(p.VoiceConnection.GuildID, p.CurrentSong)

	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tickerDone := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if p.VoiceConnection != nil && p.StreamingSession != nil && p.CurrentSong != nil {
					if !p.StreamingSession.Paused() {
						err := history.AddPlaybackDurationStats(p.VoiceConnection.GuildID, p.CurrentSong.ID, float64(interval.Seconds()))
						if err != nil {
							slog.Warnf("Error adding playback duration stats to history: %v", err)
						}
					}
				}
			case <-tickerDone:
				return
			}
		}
	}()

	select {
	case <-done:
		if p.VoiceConnection != nil && p.StreamingSession != nil && p.CurrentSong != nil {
			songDuration, songPosition := p.metrics(encodingSession, p.StreamingSession, p.CurrentSong)
			if p.CurrentStatus == StatusPlaying && encodingSession.Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 {
				if songPosition < songDuration {
					slog.Warn("Song is done but still unfinished. Restarting from interrupted position...")

					encodingSession.Cleanup()
					p.VoiceConnection.Speaking(false)
					go p.Play(int(songPosition.Seconds()), p.CurrentSong)

					return
				}
			}
		}

		err = history.AddPlaybackCountStats(p.VoiceConnection.GuildID, p.CurrentSong.ID)
		if err != nil {
			slog.Warnf("Error adding stats count stats to history: %v", err)
		}

		if err != nil && err != io.EOF {
			slog.Warnf("Song is done but an unexpected error occurred: %v", err)

			time.Sleep(250 * time.Millisecond)
			if p.VoiceConnection != nil {
				p.VoiceConnection.Speaking(false)
			}
			p.CurrentStatus = StatusError
			encodingSession.Cleanup()

			return
		}

		slog.Info("Song is done")

		if len(p.SongQueue) == 0 {
			slog.Info("Queue is done")

			time.Sleep(250 * time.Millisecond)
			p.Stop()
			p.CurrentStatus = StatusResting

			return
		}

		time.Sleep(250 * time.Millisecond)

		slog.Info("Playing next song in queue")
		p.Play(0, nil)

	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted for skip, stopping playback")

		if p.VoiceConnection != nil {
			p.VoiceConnection.Speaking(false)
		}
		encodingSession.Cleanup()

		return
	}
}

// metrics calculates playback metrics for a song.
func (p *Player) metrics(encoding *dca.EncodeSession, streaming *dca.StreamingSession, song *Song) (songDuration, songPosition time.Duration) {
	encodingDuration := encoding.Stats().Duration
	encodingStartTime := time.Duration(encoding.Options().StartTime) * time.Second

	streamingPosition := streaming.PlaybackPosition()
	delay := encodingDuration - streamingPosition

	duration, _, _, err := parseVideoParamsFromYouTubeURL(song.DownloadURL)
	if err != nil {
		slog.Warnf("Failed to parse download URL parameters: %v", err)
	}
	songDuration = time.Duration(duration) * time.Second
	songPosition = encodingStartTime + streamingPosition + delay

	slog.Infof("Total duration: %s, Stopped at: %s", songDuration, songPosition)
	slog.Infof("Encoding ahead of streaming: %s, Encoding started time: %s", delay, encodingStartTime)

	return songDuration, songPosition
}

// GetStatus returns the current playback status.
func (p *Player) GetCurrentStatus() Status {
	return p.CurrentStatus
}

// SetStatus sets the playback status.
func (p *Player) SetCurrentStatus(status Status) {
	p.CurrentStatus = status
}

// GetSongQueue returns the song queue.
func (p *Player) GetSongQueue() []*Song {
	return p.SongQueue
}

// GetVoiceConnection returns the voice connection.
func (p *Player) GetVoiceConnection() *discordgo.VoiceConnection {
	return p.VoiceConnection
}

// SetVoiceConnection sets the voice connection.
func (p *Player) SetVoiceConnection(voiceConnection *discordgo.VoiceConnection) {
	p.VoiceConnection = voiceConnection
}

// GetCurrentSong returns the current song being played.
func (p *Player) GetCurrentSong() *Song {
	return p.CurrentSong
}

// GetStreamingSession returns the current streaming session.
func (p *Player) GetStreamingSession() *dca.StreamingSession {
	return p.StreamingSession
}
