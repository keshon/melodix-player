package player

import (
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/mod-music/history"
	"github.com/keshon/melodix-discord-player/mod-music/pkg/dca"
	"github.com/keshon/melodix-discord-player/mod-music/utils"
)

func (p *Player) Play(startAt int, song *Song) error {
	// Get current song (from queue / as arg)
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
	//defer p.GetEncodingSession().Cleanup()

	// Connect to Discord channel and be ready
	// https://github.com/bwmarrin/discordgo/issues/1357
	voiceConnection, ok := p.GetDiscordSession().VoiceConnections[p.GetGuildID()]
	if !ok {
		slog.Warn("No voice connection found. Attempting to join voice channel")
		voiceConnection, err = p.GetDiscordSession().ChannelVoiceJoin(p.GetGuildID(), p.GetChannelID(), true, false)
		if err != nil {
			return fmt.Errorf("failed to join voice channel: %w", err)
		}
	}

	slog.Info("Found voice connection", voiceConnection.ChannelID)
	// voiceConnection.ChangeChannel(p.GetChannelID(), false, false)

	slog.Info("Setting it as active", voiceConnection.ChannelID)
	p.SetVoiceConnection(voiceConnection)
	// defer p.GetVoiceConnection().Close()

	err = p.GetVoiceConnection().Speaking(true)
	if err != nil {
		slog.Error("Failed to start speaking in existing voice connection", err)
		slog.Info("Attempting for a new join to voice channel and setting it as active")
		voiceConnection, err = p.GetDiscordSession().ChannelVoiceJoin(p.GetGuildID(), p.GetChannelID(), true, false)
		if err != nil {
			return fmt.Errorf("failed to join voice channel: %w", err)
		}

		p.SetVoiceConnection(voiceConnection)

		err = p.GetVoiceConnection().Speaking(true)
		if err != nil {
			return fmt.Errorf("failed to start speaking after two attempts: %w", err)
		}
	}
	// defer p.GetVoiceConnection().Speaking(false)

	// Send encoding to Discord stream
	done := make(chan error, 1)
	stream := dca.NewStream(p.GetEncodingSession(), p.GetVoiceConnection(), done)
	p.SetStreamingSession(stream)
	p.SetCurrentStatus(StatusPlaying)

	// Add song playback duration to history
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

	// Handle signals (done / skip / stop)
	select {
	case <-done:
		slog.Info("Song is interrupted due to done signal")

		p.SetCurrentStatus(StatusResting)

		if err != nil && err != io.EOF {
			time.Sleep(250 * time.Millisecond)
			if p.GetVoiceConnection() != nil {
				p.GetVoiceConnection().Speaking(false)
			}
			p.GetEncodingSession().Cleanup()

			return fmt.Errorf("unexpected error occurred: %w", err)
		}

		if p.GetVoiceConnection() == nil {
			slog.Warn("VoiceConnection is nil")
			return nil
		}

		if p.GetStreamingSession() == nil {
			slog.Error("StreamingSession is nil")
			//return nil
		}

		if p.GetCurrentSong() != nil {

			if p.GetCurrentSong().Source != SourceStream {
				// Treat as a song
				slog.Info("Checking for song metrics if interruption was unintentional")

				songDuration, songPosition, err := p.calculateSongMetrics(p.GetEncodingSession(), p.GetStreamingSession(), p.GetCurrentSong())

				if err != nil {
					return fmt.Errorf("error getting song metrics: %w", err)
				}

				if p.GetEncodingSession().Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
					startAt := songPosition.Seconds()

					p.GetEncodingSession().Cleanup()
					p.GetVoiceConnection().Speaking(false)

					slog.Warnf("Interruption detected, restarting song %v from %v", p.GetCurrentSong().Title, int(startAt))
					go func() {
						err := p.Play(int(startAt), p.GetCurrentSong())
						if err != nil {
							slog.Errorf("error restarting song: %w", err)
						}
					}()

					return nil
				}

			} else {
				// Treat as a stream
				slog.Info("Song is a stream, should always restart")

				p.GetEncodingSession().Cleanup()
				p.GetVoiceConnection().Speaking(false)

				slog.Infof("Restarting stream %v", p.GetCurrentSong().Title)
				go func() {
					err := p.Play(int(0), p.GetCurrentSong())
					if err != nil {
						slog.Errorf("error restarting song: %w", err)
					}
				}()
			}

			time.Sleep(250 * time.Millisecond)
			if err := h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID); err != nil {
				return fmt.Errorf("error adding playback count stats to history: %v", err)
			}
		}

		if p.GetCurrentSong() == nil {
			slog.Warn("CurrentSong is nil at this point")
		} else {
			slog.Warn("CurrentSong is NOT nil at this point")
		}

		p.GetVoiceConnection().Speaking(false)
		// p.GetEncodingSession().Cleanup()

		if len(p.GetSongQueue()) == 0 {
			time.Sleep(250 * time.Millisecond)
			slog.Info("Audio done")

			p.GetVoiceConnection().Speaking(false)
			p.GetVoiceConnection().Disconnect()

			p.GetEncodingSession().Cleanup()
			if p.GetVoiceConnection() != nil {
				p.GetVoiceConnection().Speaking(false)
				p.SetStreamingSession(nil)
			}

			p.SetCurrentStatus(StatusResting)
			p.SetSongQueue(make([]*Song, 0))
			p.SetCurrentSong(nil)
			p.SkipInterrupt = make(chan bool, 1)
			p.StopInterrupt = make(chan bool, 1)
			p.SwitchChannelInterrupt = make(chan bool, 1)

			return nil
		}

		time.Sleep(250 * time.Millisecond)
		slog.Warn("Play next in queue after all signals passed")
		go func() {
			err := p.Play(0, nil)
			if err != nil {
				slog.Error("Error: ", err)
			}
		}()

		slog.Info("..finished processing done signal")
		return nil
	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted due to skip signal")

		p.SetCurrentStatus(StatusResting)
		p.GetEncodingSession().Cleanup()
		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
		}

		slog.Info("..finished processing skip signal")
		return nil
	case <-p.StopInterrupt:
		slog.Info("Song is interrupted due to stop signal")

		p.GetVoiceConnection().Speaking(false)
		p.GetVoiceConnection().Disconnect()

		p.GetEncodingSession().Cleanup()
		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
			p.SetStreamingSession(nil)
		}

		p.SetCurrentStatus(StatusResting)
		p.SetSongQueue(make([]*Song, 0))
		p.SetCurrentSong(nil)
		p.SkipInterrupt = make(chan bool, 1)
		p.StopInterrupt = make(chan bool, 1)
		p.SwitchChannelInterrupt = make(chan bool, 1)

		slog.Info("..finish processing stop signal")
		return nil

	case <-p.SwitchChannelInterrupt:
		slog.Info("Song is interrupted due to switch channel signal")

		p.GetVoiceConnection().Disconnect()
		p.GetEncodingSession().Cleanup()

		go p.Play(0, p.GetCurrentSong())

		slog.Info("..finish processing switch channel signal")
		return nil
	}

}

func (p *Player) createEncodeOptions(startAt int) (*dca.EncodeOptions, error) {
	config, err := config.NewConfig()
	if err != nil {
		return nil, fmt.Errorf("error loading config: %w", err)
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

	return options, nil
}

func (p *Player) calculateSongMetrics(encodingSession *dca.EncodeSession, streamingSession *dca.StreamingSession, song *Song) (duration, position time.Duration, err error) {
	encodingDuration := encodingSession.Stats().Duration
	encodingStartTime := time.Duration(encodingSession.Options().StartTime) * time.Second

	streamingPosition := streamingSession.PlaybackPosition()
	delay := encodingDuration - streamingPosition

	parsedURL, err := url.Parse(song.DownloadURL)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse URL: %v", err)
	}
	queryParams := parsedURL.Query()

	dur, err := utils.ParseFloat64(queryParams.Get("dur"))
	if err != nil {
		return 0, 0, fmt.Errorf("failed to parse duration: %v", err)
	}

	duration, err = time.ParseDuration(fmt.Sprintf("%vs", dur))
	if err != nil {
		return 0, 0, err
	}
	position = encodingStartTime + streamingPosition + delay.Abs() // delay is negative so we make it positive to jump ahead

	slog.Debugf("Total song duration: %s, Stopped at: %s", duration, position)
	slog.Debugf("Encoding ahead of streaming: %s, Encoding started time: %s", delay, encodingStartTime)

	return duration, position, nil
}
