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

// Play starts playing audio from the given start position or song.
func (p *Player) Play(startAt int, song *Song) error {
	// Get current song (from queue or as arg)
	var currentSong *Song
	if song == nil {
		var err error
		currentSong, err = p.Dequeue()
		if err != nil {
			slog.Error(err)
			return fmt.Errorf("failed to dequeue song: %w", err)
		}
	} else {
		currentSong = song
	}
	p.SetCurrentSong(currentSong)

	// Setup encoding
	options, err := p.createEncodeOptions(startAt)
	if err != nil {
		slog.Errorf("Failed to create encode options: %v", err)
		return fmt.Errorf("failed to create encode options: %w", err)
	}

	// Start encoding
	encoding, err := dca.EncodeFile(p.GetCurrentSong().DownloadURL, options)
	if err != nil {
		slog.Error(err)
		return fmt.Errorf("failed to encode file: %w", err)
	}
	p.SetEncodingSession(encoding)
	defer p.GetEncodingSession().Cleanup()

	// Connect to Discord channel and be ready
	p.setupVoiceConnection()

	// Send encoding to Discord stream
	done := make(chan error, 1)
	stream := dca.NewStream(p.GetEncodingSession(), p.GetVoiceConnection(), done)
	p.SetStreamingSession(stream)

	// Set player status
	p.SetCurrentStatus(StatusPlaying)

	// Add current track to history
	// Add current song playback duration stat to history
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

	// Handle signals
	select {
	case <-done:
		if err != nil && err != io.EOF {
			slog.Errorf("Song is done but an unexpected error occurred: %v", err)

			time.Sleep(250 * time.Millisecond)
			if p.GetVoiceConnection() != nil {
				p.GetVoiceConnection().Speaking(false)
			}
			p.SetCurrentStatus(StatusResting)
			p.GetEncodingSession().Cleanup()

			return fmt.Errorf("unexpected error occurred: %w", err)
		}

		if p.GetVoiceConnection() == nil {
			slog.Warn("VoiceConnection is nil")
			return nil
		}

		if p.GetStreamingSession() == nil {
			slog.Warn("StreamingSession is nil")
			return nil
		}

		if p.GetCurrentSong() == nil {
			slog.Warn("CurrentSong is nil")
			return nil
		}

		if p.GetCurrentSong().Source != SourceStream {
			// handle song
			slog.Warn("Song got done signal, checking if should restart")

			songDuration, songPosition, err := p.getSongMetrics(p.GetEncodingSession(), p.GetStreamingSession(), p.GetCurrentSong())

			if err != nil {
				slog.Warnf("Error getting song metrics: %v", err)
				return fmt.Errorf("error getting song metrics: %w", err)
			}

			if p.GetEncodingSession().Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
				startAt := songPosition.Seconds()

				p.GetEncodingSession().Cleanup()
				p.GetVoiceConnection().Speaking(false)

				slog.Warnf("Restarting song %v from interrupted position %v", p.GetCurrentSong().Title, int(startAt))
				p.Play(int(startAt), p.GetCurrentSong())
			}
		} else {
			// handle stream
			slog.Warn("Stream got done signal, should always restart")

			p.GetEncodingSession().Cleanup()
			p.GetVoiceConnection().Speaking(false)

			slog.Warnf("Restarting stream %v", p.GetCurrentSong().Title)
			p.Play(0, p.GetCurrentSong())
		}

		// h := history.NewHistory()

		time.Sleep(250 * time.Millisecond)

		if err := h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID); err != nil {
			return fmt.Errorf("error adding stats count stats to history: %v", err)
		}
	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted for skip, stopping playback")

		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
		}
		p.GetEncodingSession().Cleanup()

		return nil

	case <-p.StopInterrupt:
		slog.Info("Song is interrupted for stop, stopping playback")

		p.GetStreamingSession().TryLock()
		p.GetStreamingSession().SetPaused(true)
		p.GetStreamingSession().Unlock()

		p.GetEncodingSession().Cleanup()

		p.SetSongQueue(make([]*Song, 0))
		p.SetCurrentStatus(StatusResting)
		p.SetCurrentSong(nil)

		p.SkipInterrupt = make(chan bool, 1)
		p.StopInterrupt = make(chan bool, 1)

		return nil
	}

	// if len(p.GetSongQueue()) == 0 {
	// 	time.Sleep(250 * time.Millisecond)
	// 	slog.Info("Audio done")
	// 	p.Stop()
	// 	p.SetCurrentStatus(StatusResting)
	// 	return nil
	// }

	// p.GetVoiceConnection().Speaking(false)
	// time.Sleep(250 * time.Millisecond)
	// slog.Info("Play next in queue")
	// go p.Play(0, nil)

	return nil
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
	for {
		vc := p.GetVoiceConnection()
		if vc == nil || !vc.Ready {
			time.Sleep(100 * time.Millisecond)
			slog.Warn("Voice connection is nil or not ready. Retrying...")
			continue
		}

		err := vc.Speaking(true)
		if err != nil {
			slog.Warnf("Error connecting to Discord voice: %v. Retrying...", err)
			vc.Speaking(false)
		}

		// Break out of the loop if the voice connection is ready
		if vc.Ready {
			break
		}
	}
}

// getSongMetrics calculates playback metrics for a song.
func (p *Player) getSongMetrics(encoding *dca.EncodeSession, streaming *dca.StreamingSession, song *Song) (songDuration, songPosition time.Duration, err error) {
	encodingDuration := encoding.Stats().Duration
	encodingStartTime := time.Duration(encoding.Options().StartTime) * time.Second

	streamingPosition := streaming.PlaybackPosition()
	delay := encodingDuration - streamingPosition

	params, err := utils.ParseQueryParamsFromURL(song.DownloadURL)
	if err != nil {
		slog.Warnf("Failed to parse download URL parameters: %v", err)
		return 0, 0, err
	}
	slog.Debug(params)

	// Convert duration string to time.Duration
	songDuration, err = time.ParseDuration(fmt.Sprintf("%vs", params.Duration)) // was 'duration'
	if err != nil {
		slog.Errorf("Error parsing duration:", err)
		return 0, 0, err
	}
	songPosition = encodingStartTime + streamingPosition + delay.Abs() // delay is negative so we make it positive to jump ahead

	slog.Infof("Total duration: %s, Stopped at: %s", songDuration, songPosition)
	slog.Infof("Encoding ahead of streaming: %s, Encoding started time: %s", delay, encodingStartTime)

	return songDuration, songPosition, nil
}
