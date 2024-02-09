package player

import (
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/music/history"
	"github.com/keshon/melodix-discord-player/music/pkg/dca"
	"github.com/keshon/melodix-discord-player/music/utils"
)

// Play starts playing the current or specified song.
func (p *Player) Play(startAt int, song *Song) {
	var cleanupDone sync.WaitGroup

	// Listen for skip signal
	if p.handleSkipSignal() {
		return
	}

	// Get current song (from queue or as arg)
	p.setupCurrentSong(startAt, song)

	// Setup encoding
	options, err := p.createEncodeOptions(startAt)
	if err != nil {
		slog.Fatalf("Failed to create encode options: %v", err)
	}

	// Start encoding
	var encodeSessionError error
	p.EncodingSession, encodeSessionError = dca.EncodeFile(p.CurrentSong.DownloadURL, options)
	defer func() {
		p.EncodingSession.Cleanup()
		cleanupDone.Wait()
	}()

	// Connect to Discord channel and be ready
	p.setupVoiceConnection()

	// Send encoding to Discord stream
	done := make(chan error, 1)
	p.StreamingSession = dca.NewStream(p.EncodingSession, p.VoiceConnection, done)

	// Set player status
	p.CurrentStatus = StatusPlaying

	// Setup history
	h := history.NewHistory()

	// Add current track to history
	p.addSongToHistory(h)

	// Add current song playback duration stat to history
	p.addSongPlaybackStats(h)

	// Handle done signal (finished or with error)
	p.handleDoneSignal(done, h, encodeSessionError, &cleanupDone)
}

func (p *Player) handleSkipSignal() bool {
	select {
	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted for skip, stopping playback")

		if p.VoiceConnection != nil {
			p.VoiceConnection.Speaking(false)
		}
		p.EncodingSession.Cleanup()

		return true
	default:
		// No skip signal, continue with playback
		return false
	}
}

func (p *Player) setupCurrentSong(startAt int, song *Song) {
	if song != nil {
		p.CurrentSong = song
	} else {
		if len(p.GetSongQueue()) > 0 {
			var err error
			p.CurrentSong, err = p.Dequeue()
			if err != nil {
				slog.Info("queue is empty")
			}
		}
	}

	if p.CurrentSong == nil {
		slog.Info("No songs in queue")
		return
	}
}

func (p *Player) createEncodeOptions(startAt int) (*dca.EncodeOptions, error) {
	config, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %v", err)
	}

	return &dca.EncodeOptions{
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
	}, nil
}

func (p *Player) setupEncodingSession(options *dca.EncodeOptions) error {
	var errEnc error
	p.EncodingSession, errEnc = dca.EncodeFile(p.CurrentSong.DownloadURL, options)
	return errEnc
}

func (p *Player) setupVoiceConnection() {
	for p.VoiceConnection == nil || !p.VoiceConnection.Ready {
		time.Sleep(100 * time.Millisecond)
	}

	err := p.VoiceConnection.Speaking(true)
	if err != nil {
		slog.Errorf("Error connecting to Discord voice: %v", err)
		p.VoiceConnection.Speaking(false)
	}
}

func (p *Player) addSongToHistory(h history.IHistory) {
	historySong := &history.Song{
		Name:        p.CurrentSong.Title,
		UserURL:     p.CurrentSong.UserURL,
		DownloadURL: p.CurrentSong.DownloadURL,
		Duration:    p.CurrentSong.Duration,
		ID:          p.CurrentSong.ID,
		Thumbnail:   history.Thumbnail(p.CurrentSong.Thumbnail),
	}
	h.AddTrackToHistory(p.VoiceConnection.GuildID, historySong)
}

func (p *Player) addSongPlaybackStats(h history.IHistory) {
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
						err := h.AddPlaybackDurationStats(p.VoiceConnection.GuildID, p.CurrentSong.ID, float64(interval.Seconds()))
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
	tickerDone <- true
}

func (p *Player) handleDoneSignal(done chan error, h history.IHistory, errEnc error, cleanupDone *sync.WaitGroup) {
	defer cleanupDone.Done()

	select {
	case <-done:
		if errEnc != nil && errEnc != io.EOF {
			p.handleError(errEnc)
			return
		}

		if p.VoiceConnection == nil || p.StreamingSession == nil || p.CurrentSong == nil {
			return
		}

		if p.CurrentSong.Source != SourceStream {
			p.handleSongCompletion()
		} else {
			p.handleUnfinishedSong(0)
		}

		p.handlePlaybackCountStats(h)
	}
}

func (p *Player) handleSongCompletion() {
	if p.CurrentStatus != StatusPlaying {
		return
	}

	songDuration, songPosition := p.getSongMetrics(p.EncodingSession, p.StreamingSession, p.CurrentSong)

	if p.EncodingSession.Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
		p.handleUnfinishedSong(int(songPosition.Seconds()))
	}
}

func (p *Player) handleUnfinishedSong(startAt int) {
	slog.Warn("Song is done but still unfinished. Restarting from interrupted position...")

	p.EncodingSession.Cleanup()
	p.VoiceConnection.Speaking(false)

	p.Play(startAt, p.CurrentSong)
}

func (p *Player) handlePlaybackCountStats(h history.IHistory) {
	if err := h.AddPlaybackCountStats(p.VoiceConnection.GuildID, p.CurrentSong.ID); err != nil {
		slog.Warnf("Error adding stats count stats to history: %v", err)
	}
}

func (p *Player) handleError(err error) {
	slog.Warnf("Song is done but an unexpected error occurred: %v", err)

	time.Sleep(250 * time.Millisecond)
	if p.VoiceConnection != nil {
		p.VoiceConnection.Speaking(false)
	}
	p.CurrentStatus = StatusResting
	p.EncodingSession.Cleanup()
}

// getSongMetrics calculates playback metrics for a song.
func (p *Player) getSongMetrics(encoding *dca.EncodeSession, streaming *dca.StreamingSession, song *Song) (songDuration, songPosition time.Duration) {
	encodingDuration := encoding.Stats().Duration
	encodingStartTime := time.Duration(encoding.Options().StartTime) * time.Second

	streamingPosition := streaming.PlaybackPosition()
	delay := encodingDuration - streamingPosition

	params, err := utils.ParseQueryParamsFromURL(song.DownloadURL)
	if err != nil {
		slog.Warnf("Failed to parse download URL parameters: %v", err)
	}

	// Convert duration string to time.Duration
	duration, err := time.ParseDuration(params["duration"])
	if err != nil {
		slog.Errorf("Error parsing duration:", err)
	}

	songDuration = time.Duration(duration) * time.Second
	songPosition = encodingStartTime + streamingPosition + delay

	slog.Infof("Total duration: %s, Stopped at: %s", songDuration, songPosition)
	slog.Infof("Encoding ahead of streaming: %s, Encoding started time: %s", delay, encodingStartTime)

	return songDuration, songPosition
}

func (p *Player) logPlayingInfo() {
	slog.Warnf("Current status: %s", p.GetCurrentStatus())

	if p.GetCurrentSong() != nil {
		slog.Warn("Current song: %s", p.GetCurrentSong().Title)
	} else {
		slog.Warn("Current song is null")
	}

	slog.Warn("Song queue:")
	for _, elem := range p.GetSongQueue() {
		slog.Warn(elem.Title)
	}
	slog.Warn("Playlist count is ", len(p.GetSongQueue()))
}
