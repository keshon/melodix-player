package player

import (
	"fmt"
	"io"
	"net/url"
	"os/exec"
	"regexp"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/mods/music/history"
	"github.com/keshon/melodix-player/mods/music/media"
	"github.com/keshon/melodix-player/mods/music/sources"
	"github.com/keshon/melodix-player/mods/music/third_party/dca"
	"github.com/keshon/melodix-player/mods/music/utils"
)

// Down below is One Big Fat Function to play a song
// The reason it's not split and is so logging verbose is due its complex logic flow

func (p *Player) Play(startAt int, song *media.Song) error {
	// Get current song (from queue / as arg)
	currentSong, err := func() (*media.Song, error) {
		if song != nil {
			return song, nil
		}

		dequedSong, err := p.Dequeue()
		if err != nil {
			return nil, fmt.Errorf("failed to dequeue song: %w", err)
		}

		return dequedSong, nil
	}()
	if err != nil {
		return fmt.Errorf("failed to get current song: %w", err)
	}

	p.SetCurrentSong(currentSong)

	// Setup and start encoding
	options, err := func(startAt int) (*dca.EncodeOptions, error) {
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
	}(startAt)
	if err != nil {
		return fmt.Errorf("failed to create encode options: %w", err)
	}

	encoding, err := dca.EncodeFile(p.GetCurrentSong().Filepath, options)
	if err != nil {
		return fmt.Errorf("failed to encode file: %w", err)
	}

	p.SetEncodingSession(encoding)
	defer p.GetEncodingSession().Cleanup()

	// Set up voice connection for sending audio
	voiceConnection, err := p.setupVoiceConnection()
	if err != nil {
		return err
	}

	slog.Info("Found voice connection and setting it as active", voiceConnection.ChannelID)
	p.SetVoiceConnection(voiceConnection)

	// Send encoding stream to voice connection
	done := make(chan error, 1)
	stream := dca.NewStream(p.GetEncodingSession(), p.GetVoiceConnection(), done)
	p.SetStreamingSession(stream)
	p.SetCurrentStatus(StatusPlaying)

	// Add song to history
	historySong := &history.Song{
		Title:     p.GetCurrentSong().Title,
		URL:       p.GetCurrentSong().URL,
		Filepath:  p.GetCurrentSong().Filepath,
		Duration:  p.GetCurrentSong().Duration,
		SongID:    p.GetCurrentSong().SongID,
		Thumbnail: history.Thumbnail(p.GetCurrentSong().Thumbnail),
		Source:    p.GetCurrentSong().Source.String(),
	}
	p.GetHistory().AddTrackToHistory(p.GetVoiceConnection().GuildID, historySong)

	if err := p.GetHistory().AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().SongID); err != nil {
		slog.Errorf("error adding playback count stats to history: %v", err)
	}

	// Set up periodic playback duration stats update to history
	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	tickerStop := make(chan bool)
	defer func() {
		ticker.Stop()
		tickerStop <- true
	}()

	go func() {
		for {
			select {
			case <-ticker.C:
				if p.GetVoiceConnection() != nil && p.GetStreamingSession() != nil && p.GetCurrentSong() != nil && !p.GetStreamingSession().Paused() {
					err := p.GetHistory().AddPlaybackDurationStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().SongID, float64(interval.Seconds()))
					if err != nil {
						slog.Warnf("Error adding playback duration stats to history: %v", err)
					}
				}
			case <-tickerStop:
				return
			}
		}
	}()

	// Handle signals (done / skip / stop)
	select {
	case errDone := <-done:
		slog.Info("Song is interrupted due to done signal")
		p.SetCurrentStatus(StatusResting)

		if errDone != nil && errDone != io.EOF { // ? Point of interest: handle EOF errors
			time.Sleep(250 * time.Millisecond)
			if p.GetVoiceConnection() != nil {
				p.GetVoiceConnection().Speaking(false)
			}

			slog.Error("Unexpected error occurred", errDone)
			slog.Info("Resetting voice connection...")

			if p.GetVoiceConnection() != nil {
				p.GetVoiceConnection().Speaking(false)
				p.GetVoiceConnection().Disconnect()
			}

			voiceConnection, err := p.setupVoiceConnection()
			if err != nil {
				return err
			}

			slog.Info("Found voice connection and setting it as active", voiceConnection.ChannelID)
			p.SetVoiceConnection(voiceConnection)
		}

		if p.GetVoiceConnection() == nil {
			slog.Warn("VoiceConnection is nil")
			return nil
		}

		if p.GetStreamingSession() == nil {
			slog.Error("StreamingSession is nil")
			//return nil // !Potentially a risk to bypass this
		}

		if p.GetCurrentSong() != nil {
			switch {
			case p.GetCurrentSong().Source == media.SourceYouTube:
				slog.Info("Source is a YouTube video, checking for song metrics if unexpected interruption")
				songDuration, songPosition, err := p.calculateSongMetrics(p.GetEncodingSession(), p.GetStreamingSession(), p.GetCurrentSong())
				if err != nil {
					return fmt.Errorf("error getting song metrics: %w", err)
				}

				if p.GetEncodingSession().Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
					startAt := songPosition.Seconds()
					p.GetVoiceConnection().Speaking(false)
					slog.Warnf("Unexpected interruption confirmed, restarting song: \"%v\" from %vs", p.GetCurrentSong().Title, int(startAt))

					go func() {
						yt := sources.NewYoutube()
						song, err = yt.FetchOneByURL(p.GetCurrentSong().URL)
						if err != nil {
							slog.Errorf("error fetching new song: %w", err)
						}

						p.SetCurrentSong(song)

						err := p.Play(int(startAt), p.GetCurrentSong())
						if err != nil {
							slog.Errorf("error restarting song: %w", err)
						}
					}()

					return nil
				}
				// fallthrough
			case p.GetCurrentSong().Source == media.SourceStream:
				slog.Info("Source is a stream, should always restart (unless manually interrupted)")
				p.GetVoiceConnection().Speaking(false)
				slog.Infof("Restarting stream %v", p.GetCurrentSong().Title)

				go func() {
					err := p.Play(int(0), p.GetCurrentSong())
					if err != nil {
						slog.Errorf("error restarting song: %w", err)
					}
				}()

				return nil

			case p.GetCurrentSong().Source == media.SourceLocalFile:
				slog.Info("Source is a local file, checking for song metrics if unexpected interruption")
				songDuration, songPosition, err := p.calculateSongMetrics(p.GetEncodingSession(), p.GetStreamingSession(), p.GetCurrentSong())
				if err != nil {
					return fmt.Errorf("error getting song metrics: %w", err)
				}

				if p.GetEncodingSession().Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 && songPosition < songDuration {
					startAt := songPosition.Seconds()
					p.GetVoiceConnection().Speaking(false)
					slog.Warnf("Unexpected interruption confirmed, restarting song: \"%v\" from %vs", p.GetCurrentSong().Title, int(startAt))

					go func() {
						err := p.Play(int(startAt), p.GetCurrentSong())
						if err != nil {
							slog.Errorf("error restarting song: %w", err)
						}
					}()

					return nil
				}
				// fallthrough
			}

		}

		if p.GetCurrentSong() == nil {
			slog.Error("CurrentSong is nil at this point (should not happen)")
		} else {
			slog.Warn("CurrentSong is NOT nil at this point:", p.GetCurrentSong().Title)
		}

		p.GetVoiceConnection().Speaking(false)

		if len(p.GetSongQueue()) == 0 {
			time.Sleep(250 * time.Millisecond)

			if p.GetVoiceConnection() != nil {
				p.GetVoiceConnection().Speaking(false)
				p.GetVoiceConnection().Disconnect()
			}

			p.SetStreamingSession(nil)

			p.SetCurrentStatus(StatusResting)
			p.SetSongQueue(make([]*media.Song, 0))
			p.SetCurrentSong(nil)
			p.SkipInterrupt = make(chan bool, 1)
			p.StopInterrupt = make(chan bool, 1)
			p.SwitchChannelInterrupt = make(chan bool, 1)

			slog.Info("Stop playing after all signals passed, audio is done")
			return nil
		}

		slog.Info("Play next in queue after all signals passed")

		time.Sleep(250 * time.Millisecond)
		go func() {
			err := p.Play(0, nil)
			if err != nil {
				slog.Error("Error playing next song after done signal: ", err)
			}
		}()

		slog.Info("..finished processing done signal")
		return nil
	case <-p.SkipInterrupt:
		slog.Info("Song is interrupted due to skip signal")

		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
		}

		p.SetCurrentStatus(StatusResting)

		slog.Info("..finished processing skip signal")
		return nil
	case <-p.StopInterrupt:
		slog.Info("Song is interrupted due to stop signal")

		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Speaking(false)
			p.GetVoiceConnection().Disconnect()
		}

		p.SetStreamingSession(nil)

		p.SetCurrentStatus(StatusResting)
		p.SetSongQueue(make([]*media.Song, 0))
		p.SetCurrentSong(nil)
		p.SkipInterrupt = make(chan bool, 1)
		p.StopInterrupt = make(chan bool, 1)
		p.SwitchChannelInterrupt = make(chan bool, 1)

		slog.Info("..finish processing stop signal")
		return nil

	case <-p.SwitchChannelInterrupt:
		slog.Info("Song is interrupted due to switch channel signal")

		if p.GetVoiceConnection() != nil {
			p.GetVoiceConnection().Disconnect()
		}

		go func() {
			err := p.Play(0, p.GetCurrentSong())
			if err != nil {
				slog.Error("Error playing next song after switch signal: ", err)
			}
		}()

		slog.Info("..finish processing switch channel signal")
		return nil
	}

}

func (p *Player) setupVoiceConnection() (*discordgo.VoiceConnection, error) {
	// Helpful: https://github.com/bwmarrin/discordgo/issues/1357
	session := p.GetDiscordSession()
	guildID, channelID := p.GetGuildID(), p.GetChannelID()

	var voiceConnection *discordgo.VoiceConnection
	var err error

	session.ShouldReconnectOnError = true

	for attempts := 0; attempts < 5; attempts++ {
		voiceConnection, err = session.ChannelVoiceJoin(guildID, channelID, false, false)
		if err == nil {
			break
		}

		if attempts > 0 {
			slog.Warn("Failed to join voice channel after multiple attempts, attempting to disconnect and reconnect next iteration")
			if voiceConnection != nil {
				voiceConnection.Disconnect()
			}
		}

		slog.Warnf("Failed to join voice channel (attempt %d): %v", attempts+1, err)
		time.Sleep(300 * time.Millisecond)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to join voice channel after multiple attempts: %w", err)
	}

	slog.Info("Successfully joined voice channel")
	return voiceConnection, nil
}

func (p *Player) calculateSongMetrics(encodingSession *dca.EncodeSession, streamingSession *dca.StreamingSession, song *media.Song) (duration, position time.Duration, err error) {
	encodingDuration := encodingSession.Stats().Duration
	encodingStartTime := time.Duration(encodingSession.Options().StartTime) * time.Second

	streamingPosition := streamingSession.PlaybackPosition()
	delay := encodingDuration - streamingPosition

	var dur float64
	switch song.Source {
	case media.SourceYouTube:
		parsedURL, err := url.Parse(song.Filepath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse URL: %v", err)
		}
		queryParams := parsedURL.Query()
		dur, err = utils.ParseFloat64(queryParams.Get("dur"))
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse duration: %v", err)
		}
	case media.SourceLocalFile:
		dur, err = getMP3Duration(song.Filepath)
		if err != nil {
			return 0, 0, fmt.Errorf("failed to parse duration: %v", err)
		}
	default:
		return 0, 0, fmt.Errorf("unknown source: %v", song.Source)
	}

	duration, err = time.ParseDuration(fmt.Sprintf("%vs", dur))
	if err != nil {
		return 0, 0, err
	}
	position = encodingStartTime + streamingPosition + delay.Abs() // delay is added and not subtracted so we won't end up stuck in a restarting loop

	slog.Debugf("Song stopped at:\t%s,\tSong duration:\t%s", position, duration)
	slog.Debugf("Encoding started at:\t%s,\tEncoding ahead:\t%s", encodingStartTime, delay)

	return duration, position, nil
}

func getMP3Duration(filePath string) (float64, error) {
	cmd := exec.Command("ffmpeg", "-i", filePath, "-f", "null", "-")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, err
	}

	// Extract duration information using regular expression
	durationRegex := regexp.MustCompile(`Duration: (\d{2}):(\d{2}):(\d{2})\.\d+`)
	matches := durationRegex.FindStringSubmatch(string(output))
	if len(matches) != 4 {
		return 0, fmt.Errorf("duration not found in ffmpeg output")
	}

	// Convert hours, minutes, and seconds to seconds and combine
	hours, _ := strconv.Atoi(matches[1])
	minutes, _ := strconv.Atoi(matches[2])
	seconds, _ := strconv.Atoi(matches[3])
	totalSeconds := float64(hours*3600 + minutes*60 + seconds)

	return totalSeconds, nil
}
