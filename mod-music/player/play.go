package player

import (
	"fmt"
	"io"
	"time"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/mod-music/history"
	"github.com/keshon/melodix-discord-player/mod-music/pkg/dca"
	"github.com/keshon/melodix-discord-player/mod-music/utils"
)

// Play starts playing the current or specified song.
func (p *Player) Play(startAt int, song *Song) {
	// Get current song (from queue or as arg)
	var currentSong *Song
	if song == nil {
		var err error
		currentSong, err = p.Dequeue()
		if err != nil {
			slog.Error(err)
			return
		}
	} else {
		currentSong = song
	}
	p.SetCurrentSong(currentSong)

	// Setup encoding
	options, err := p.createEncodeOptions(startAt)
	if err != nil {
		slog.Fatalf("Failed to create encode options: %v", err)
		return
	}

	// Start encoding
	encodingSession, encodingError := dca.EncodeFile(p.GetCurrentSong().DownloadURL, options)
	if encodingError != nil {
		slog.Error(encodingError)
		return
	}
	p.SetEncodingSession(encodingSession)
	defer p.GetEncodingSession().Cleanup()

	// Connect to Discord channel and be ready
	p.setupVoiceConnection()

	// Send encoding to Discord stream
	done := make(chan error, 1)
	streamSession := dca.NewStream(p.GetEncodingSession(), p.GetVoiceConnection(), done)
	p.SetStreamingSession(streamSession)

	// Set player status
	p.SetCurrentStatus(StatusPlaying)

	// Add current track to history
	p.addSongToHistory()

	// Add current song playback duration stat to history
	p.addSongPlaybackStats()

	// Handle signals
	select {
	case <-done:
		if encodingError != nil && encodingError != io.EOF {
			p.handleSongError(encodingError)
			return
		}

		if p.GetVoiceConnection() == nil || p.GetStreamingSession() == nil || p.GetCurrentSong() == nil {
			return
		}

		if p.GetCurrentSong().Source != SourceStream {
			p.handleSongCompletion()
		} else {
			p.handleSongUnfinished(0)
		}

		h := history.NewHistory()
		if err := h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID); err != nil {
			slog.Warnf("Error adding stats count stats to history: %v", err)
		}
	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted for skip, stopping playback")

		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
		}
		p.GetEncodingSession().Cleanup()
		return
	}

	if len(p.GetSongQueue()) == 0 {
		time.Sleep(250 * time.Millisecond)
		slog.Info("Audio done")
		p.Stop()
		p.SetCurrentStatus(StatusResting)
		return
	}

	p.GetVoiceConnection().Speaking(false)
	time.Sleep(250 * time.Millisecond)
	slog.Info("Play next in queue")
	go p.Play(0, nil)
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

func (p *Player) addSongToHistory() {
	historySong := &history.Song{
		Name:        p.GetCurrentSong().Title,
		UserURL:     p.GetCurrentSong().UserURL,
		DownloadURL: p.GetCurrentSong().DownloadURL,
		Duration:    p.GetCurrentSong().Duration,
		ID:          p.GetCurrentSong().ID,
		Thumbnail:   history.Thumbnail(p.GetCurrentSong().Thumbnail),
	}
	h := history.NewHistory()
	h.AddTrackToHistory(p.GetVoiceConnection().GuildID, historySong)
}

func (p *Player) addSongPlaybackStats() {
	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	tickerDone := make(chan bool)

	h := history.NewHistory()

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

func (p *Player) handleSongCompletion() {
	if p.GetCurrentStatus() != StatusPlaying {
		return
	}

	songDuration, songPosition := p.getSongMetrics(p.GetEncodingSession(), p.GetStreamingSession(), p.GetCurrentSong())

	if p.GetEncodingSession().Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
		startAt := songPosition.Seconds()
		slog.Warn("Track should be continue starting from ", startAt)
		p.handleSongUnfinished(int(startAt))
	}
}

func (p *Player) handleSongUnfinished(startAt int) {
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

func (p *Player) handleSongError(err error) {
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
