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
	var encErr error
	encSession, encErr := dca.EncodeFile(p.GetCurrentSong().DownloadURL, options)
	p.SetEncodingSession(encSession)
	defer func() {
		p.GetEncodingSession().Cleanup()
		cleanupDone.Wait()
	}()

	// Connect to Discord channel and be ready
	p.setupVoiceConnection()

	// Send encoding to Discord stream
	done := make(chan error, 1)
	streamSession := dca.NewStream(p.GetEncodingSession(), p.GetVoiceConnection(), done)
	p.SetStreamingSession(streamSession)

	// Set player status
	p.SetCurrentStatus(StatusPlaying)

	// Setup history
	h := history.NewHistory()

	// Add current track to history
	p.addSongToHistory(h)

	// Add current song playback duration stat to history
	p.addSongPlaybackStats(h)

	// Handle done signal (finished or with error)
	p.handleDoneSignal(done, h, encErr, &cleanupDone)
}

func (p *Player) handleSkipSignal() bool {
	select {
	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted for skip, stopping playback")

		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
		}
		p.GetEncodingSession().Cleanup()
		// p.SetCurrentStatus(StatusResting)

		return true
	default:
		// No skip signal, continue with playback
		return false
	}
}

func (p *Player) setupCurrentSong(startAt int, song *Song) {
	if song != nil {
		p.SetCurrentSong(song)
	} else {
		if len(p.GetSongQueue()) > 0 {
			var err error
			currentSong, err := p.Dequeue()
			if err != nil {
				slog.Info("Error dequening: ", err)
			} else {
				p.SetCurrentSong(currentSong)
			}
		}
	}

	if p.GetCurrentSong() == nil {
		slog.Info("No songs in queue")
		return
	}
}

func (p *Player) createEncodeOptions(startAt int) (*dca.EncodeOptions, error) {
	config, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
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
	session, errEnc := dca.EncodeFile(p.GetCurrentSong().DownloadURL, options)
	p.SetEncodingSession(session)
	return errEnc
}

func (p *Player) setupVoiceConnection() {
	// p.Lock()
	// defer p.Unlock()

	for p.GetVoiceConnection() == nil || !p.GetVoiceConnection().Ready {
		time.Sleep(100 * time.Millisecond)

		if p.GetVoiceConnection() == nil {
			slog.Warn("Voice connection is nil. Retrying...")
			continue
		}

		err := p.GetVoiceConnection().Speaking(true)
		if err != nil {
			slog.Warnf("Error connecting to Discord voice: %v. Retrying...", err)
			p.GetVoiceConnection().Speaking(false)
		}
	}
}

func (p *Player) addSongToHistory(h history.IHistory) {
	historySong := &history.Song{
		Name:        p.GetCurrentSong().Title,
		UserURL:     p.GetCurrentSong().UserURL,
		DownloadURL: p.GetCurrentSong().DownloadURL,
		Duration:    p.GetCurrentSong().Duration,
		ID:          p.GetCurrentSong().ID,
		Thumbnail:   history.Thumbnail(p.GetCurrentSong().Thumbnail),
	}
	h.AddTrackToHistory(p.GetVoiceConnection().GuildID, historySong)
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
				if p.GetVoiceConnection() != nil && p.GetStreamingSession() != nil && p.GetCurrentSong() != nil {
					if !p.GetStreamingSession().Paused() {
						err := h.AddPlaybackDurationStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID, float64(interval.Seconds()))
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

		if p.GetVoiceConnection() == nil || p.GetStreamingSession() == nil || p.GetCurrentSong() == nil {
			return
		}

		if p.GetCurrentSong().Source != SourceStream {
			p.handleSongCompletion()
		} else {
			p.handleUnfinishedSong(0)
		}

		p.handlePlaybackCountStats(h)
	}
}

func (p *Player) handleSongCompletion() {
	if p.GetCurrentStatus() != StatusPlaying {
		return
	}

	songDuration, songPosition := p.getSongMetrics(p.GetEncodingSession(), p.GetStreamingSession(), p.GetCurrentSong())

	if p.GetEncodingSession().Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
		startAt := songPosition.Seconds()
		slog.Warn("Track should be continue starting from ", startAt)
		p.handleUnfinishedSong(int(startAt))
	}
}

func (p *Player) handleUnfinishedSong(startAt int) {
	slog.Warn("Song is done but still unfinished. Restarting from interrupted position...")

	p.GetEncodingSession().Cleanup()
	p.GetVoiceConnection().Speaking(false)

	if p.GetCurrentSong() != nil {
		slog.Errorf("Current song should not be empty", p.GetCurrentSong())
		p.Play(startAt, p.GetCurrentSong())
	} else {
		slog.Warn("No songs in the queue to restart from.")
		p.SetCurrentStatus(StatusResting)
	}
}

func (p *Player) handlePlaybackCountStats(h history.IHistory) {
	if err := h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID); err != nil {
		slog.Warnf("Error adding stats count stats to history: %v", err)
	}
}

func (p *Player) handleError(err error) {
	slog.Warnf("Song is done but an unexpected error occurred: %v", err)

	time.Sleep(250 * time.Millisecond)
	if p.GetVoiceConnection() != nil {
		p.GetVoiceConnection().Speaking(false)
	}
	p.SetCurrentStatus(StatusResting)
	p.GetEncodingSession().Cleanup()
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
	slog.Debug(params)

	// Convert duration string to time.Duration
	songDuration, err = time.ParseDuration(fmt.Sprintf("%vs", params.Duration)) // was 'duration'
	if err != nil {
		slog.Errorf("Error parsing duration:", err)
	}
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
