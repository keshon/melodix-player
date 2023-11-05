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

// Status represents the playback status of the MelodixPlayer.
type Status int32

const (
	StatusResting Status = iota
	StatusPlaying
	StatusPaused
	StatusError
)

// MelodixPlayer manages audio playback and song queue.
type MelodixPlayer struct {
	sync.Mutex
	VoiceConnection  *discordgo.VoiceConnection
	StreamingSession *dca.StreamingSession
	SongQueue        []*Song
	CurrentSong      *Song
	CurrentStatus    Status
	SkipInterrupt    chan bool
}

// IMelodixPlayer defines the interface for managing audio playback and song queue.
type IMelodixPlayer interface {
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

// NewPlayer creates a new MelodixPlayer instance.
func NewPlayer(guildID string) IMelodixPlayer {
	return &MelodixPlayer{
		SongQueue:     make([]*Song, 0),
		SkipInterrupt: make(chan bool, 1),
		CurrentStatus: StatusResting,
	}
}

// Skip skips to the next song in the queue.
func (mp *MelodixPlayer) Skip() {
	slog.Info("Skipping to next song")

	if mp.VoiceConnection == nil {
		return
	}

	mp.CurrentStatus = StatusResting

	if len(mp.SkipInterrupt) == 0 {
		history := NewHistory()
		history.AddPlaybackCountStats(mp.VoiceConnection.GuildID, mp.CurrentSong.ID)

		mp.SkipInterrupt <- true
		mp.Play(0, nil)
	}
}

// Enqueue adds a song to the queue.
func (mp *MelodixPlayer) Enqueue(song *Song) {
	slog.Infof("Enqueuing song to queue: %v", song.Name)

	mp.Lock()
	defer mp.Unlock()

	mp.SongQueue = append(mp.SongQueue, song)
}

// Dequeue removes and returns the first song from the queue.
func (mp *MelodixPlayer) Dequeue() *Song {
	slog.Info("Dequeuing song and returning it from queue")

	mp.Lock()
	defer mp.Unlock()

	if len(mp.SongQueue) == 0 {
		return nil
	}

	firstSong := mp.SongQueue[0]
	mp.SongQueue = mp.SongQueue[1:]

	return firstSong
}

// ClearQueue clears the song queue.
func (mp *MelodixPlayer) ClearQueue() {
	slog.Info("Clearing song queue")
	mp.SongQueue = make([]*Song, 0)
}

// Stop stops audio playback and disconnects from the voice channel.
func (mp *MelodixPlayer) Stop() {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	mp.CurrentStatus = StatusResting

	if mp.VoiceConnection == nil {
		return
	}

	err := mp.VoiceConnection.Disconnect()
	if err != nil {
		slog.Errorf("Error disconnecting voice connection: %v", err)
	}

	mp.VoiceConnection = nil
	mp.StreamingSession = nil
	mp.CurrentSong = nil
}

// Pause pauses audio playback.
func (mp *MelodixPlayer) Pause() {
	slog.Info("Pausing audio playback")

	if mp.VoiceConnection == nil {
		return
	}

	if mp.StreamingSession == nil {
		return
	}

	if mp.CurrentStatus == StatusPlaying {
		mp.StreamingSession.SetPaused(true)
		mp.CurrentStatus = StatusPaused
	}
}

// Unpause resumes audio playback.
func (mp *MelodixPlayer) Unpause() {
	slog.Info("Resuming playback")

	if mp.StreamingSession != nil {
		if mp.CurrentStatus != StatusPlaying {
			mp.StreamingSession.SetPaused(false)
			mp.CurrentStatus = StatusPlaying
		}
	} else {
		mp.Play(0, nil)
		mp.CurrentStatus = StatusPlaying
	}
}

// Play starts playing the current or specified song.
func (mp *MelodixPlayer) Play(startAt int, song *Song) {

	if song == nil {
		mp.CurrentSong = mp.Dequeue()
		if mp.CurrentSong == nil {
			slog.Info("No songs in queue")
			mp.CurrentStatus = StatusResting
			return
		}
	}

	slog.Infof("Playing song: %v", mp.CurrentSong.Name)
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

	encodingSession, err := dca.EncodeFile(mp.CurrentSong.DownloadURL, options)
	if err != nil {
		slog.Errorf("Error encoding song: %v", err)
		return
	}
	defer encodingSession.Cleanup()

	for mp.VoiceConnection == nil || !mp.VoiceConnection.Ready {
		time.Sleep(100 * time.Millisecond)
	}

	err = mp.VoiceConnection.Speaking(true)
	if err != nil {
		slog.Errorf("Error connecting to Discord voice: %v", err)
		mp.VoiceConnection.Speaking(false)
		return
	}

	done := make(chan error)
	mp.StreamingSession = dca.NewStream(encodingSession, mp.VoiceConnection, done)

	slog.Info("Stream is created, waiting for finish or error")

	mp.CurrentStatus = StatusPlaying

	history := NewHistory()
	history.AddTrackToHistory(mp.VoiceConnection.GuildID, mp.CurrentSong)

	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tickerDone := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				if mp.VoiceConnection != nil && mp.StreamingSession != nil && mp.CurrentSong != nil {
					if !mp.StreamingSession.Paused() {
						err := history.AddPlaybackDurationStats(mp.VoiceConnection.GuildID, mp.CurrentSong.ID, float64(interval.Seconds()))
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
		if mp.VoiceConnection != nil && mp.StreamingSession != nil && mp.CurrentSong != nil {
			songDuration, songPosition := mp.metrics(encodingSession, mp.StreamingSession, mp.CurrentSong)
			if mp.CurrentStatus == StatusPlaying && encodingSession.Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 {
				if songPosition < songDuration {
					slog.Warn("Song is done but still unfinished. Restarting from interrupted position...")

					encodingSession.Cleanup()
					mp.VoiceConnection.Speaking(false)
					go mp.Play(int(songPosition.Seconds()), mp.CurrentSong)

					return
				}
			}
		}

		err = history.AddPlaybackCountStats(mp.VoiceConnection.GuildID, mp.CurrentSong.ID)
		if err != nil {
			slog.Warnf("Error adding stats count stats to history: %v", err)
		}

		if err != nil && err != io.EOF {
			slog.Warnf("Song is done but an unexpected error occurred: %v", err)

			time.Sleep(250 * time.Millisecond)
			if mp.VoiceConnection != nil {
				mp.VoiceConnection.Speaking(false)
			}
			mp.CurrentStatus = StatusError
			encodingSession.Cleanup()

			return
		}

		slog.Info("Song is done")

		if len(mp.SongQueue) == 0 {
			slog.Info("Queue is done")

			time.Sleep(250 * time.Millisecond)
			mp.Stop()
			mp.CurrentStatus = StatusResting

			return
		}

		time.Sleep(250 * time.Millisecond)

		slog.Info("Playing next song in queue")
		mp.Play(0, nil)

	case <-mp.SkipInterrupt:
		slog.Info("Song is interrupted for skip, stopping playback")

		if mp.VoiceConnection != nil {
			mp.VoiceConnection.Speaking(false)
		}
		encodingSession.Cleanup()

		return
	}
}

// metrics calculates playback metrics for a song.
func (mp *MelodixPlayer) metrics(encoding *dca.EncodeSession, streaming *dca.StreamingSession, song *Song) (songDuration, songPosition time.Duration) {
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
func (mp *MelodixPlayer) GetCurrentStatus() Status {
	return mp.CurrentStatus
}

// SetStatus sets the playback status.
func (mp *MelodixPlayer) SetCurrentStatus(status Status) {
	mp.CurrentStatus = status
}

// GetSongQueue returns the song queue.
func (mp *MelodixPlayer) GetSongQueue() []*Song {
	return mp.SongQueue
}

// GetVoiceConnection returns the voice connection.
func (mp *MelodixPlayer) GetVoiceConnection() *discordgo.VoiceConnection {
	return mp.VoiceConnection
}

// SetVoiceConnection sets the voice connection.
func (mp *MelodixPlayer) SetVoiceConnection(voiceConnection *discordgo.VoiceConnection) {
	mp.VoiceConnection = voiceConnection
}

// GetCurrentSong returns the current song being played.
func (mp *MelodixPlayer) GetCurrentSong() *Song {
	return mp.CurrentSong
}

// GetStreamingSession returns the current streaming session.
func (mp *MelodixPlayer) GetStreamingSession() *dca.StreamingSession {
	return mp.StreamingSession
}
