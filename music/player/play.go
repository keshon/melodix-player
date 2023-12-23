package player

import (
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
	options := p.createEncodeOptions(startAt)

	// Start encoding
	var encodeSessionError error
	p.EncodingSession, encodeSessionError = dca.EncodeFile(p.CurrentSong.DownloadURL, options)
	defer p.EncodingSession.Cleanup()

	// Connect to Discord channel and be ready
	p.setupVoiceConnection()

	// Send encoding to Discord stream
	done := make(chan error)
	p.StreamingSession = dca.NewStream(p.EncodingSession, p.VoiceConnection, done)

	// Set player status
	p.CurrentStatus = StatusPlaying

	// Setup history
	h := history.NewHistory()

	// Add current track to history
	p.addSongToHistory(h)

	p.setupPlaybackDurationStatsTicker(h)

	// Done signal
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
			p.CurrentSong = p.Dequeue()
		}
	}

	if p.CurrentSong == nil {
		slog.Info("No songs in queue")
		return
	}
}

func (p *Player) createEncodeOptions(startAt int) *dca.EncodeOptions {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
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
	}
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

func (p *Player) setupPlaybackDurationStatsTicker(h history.IHistory) {
	interval := 2 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	tickerDone := make(chan bool)

	go func() {
		for {
			select {
			case <-ticker.C:
				p.addPlaybackStatsToHistory(h, interval)
			case <-tickerDone:
				return
			}
		}
	}()
}

func (p *Player) addPlaybackStatsToHistory(h history.IHistory, interval time.Duration) {
	if p.VoiceConnection != nil && p.StreamingSession != nil && p.CurrentSong != nil {
		if !p.StreamingSession.Paused() {
			err := h.AddPlaybackDurationStats(p.VoiceConnection.GuildID, p.CurrentSong.ID, float64(interval.Seconds()))
			if err != nil {
				slog.Warnf("Error adding playback duration stats to history: %v", err)
			}
		}
	}
}

func (p *Player) handleDoneSignal(done chan error, h history.IHistory, errEnc error, cleanupDone *sync.WaitGroup) {
	select {
	case <-done:
		cleanupDone.Add(1)
		go func() {
			// Auto-restarting logic in case of interruption
			// Youtube songs checked by their current vs total duration
			// Streams (radio) never stop
			if p.VoiceConnection != nil && p.StreamingSession != nil && p.CurrentSong != nil {
				if p.CurrentSong.Source != SourceStream {
					songDuration, songPosition := p.getSongMetrics(p.EncodingSession, p.StreamingSession, p.CurrentSong)
					if p.CurrentStatus == StatusPlaying {
						if p.EncodingSession.Stats().Duration.Seconds() > 0 && songPosition.Seconds() > 0 {
							if songPosition < songDuration {
								slog.Warn("Song is done but still unfinished. Restarting from interrupted position...")

								p.EncodingSession.Cleanup()
								p.VoiceConnection.Speaking(false)

								p.Play(int(songPosition.Seconds()), p.CurrentSong)

								return
							}
						}
					}
				} else {
					if p.CurrentStatus == StatusPlaying {

						slog.Warn("Song is done but its a stream so it's never finished. Restarting from interrupted position...")

						p.EncodingSession.Cleanup()
						p.VoiceConnection.Speaking(false)

						p.Play(0, p.CurrentSong)

						return

					}
				}

				err := h.AddPlaybackCountStats(p.VoiceConnection.GuildID, p.CurrentSong.ID)
				if err != nil {
					slog.Warnf("Error adding stats count stats to history: %v", err)
				}
			}

			if errEnc != nil && errEnc != io.EOF {
				slog.Warnf("Song is done but an unexpected error occurred: %v", errEnc)

				time.Sleep(250 * time.Millisecond)
				if p.VoiceConnection != nil {
					p.VoiceConnection.Speaking(false)
				}
				p.CurrentStatus = StatusResting
				p.EncodingSession.Cleanup()

				return
			}

			slog.Info("Song is done")

			if len(p.GetSongQueue()) == 0 {
				slog.Info("Queue is done")

				time.Sleep(250 * time.Millisecond)
				p.Stop()

				return
			}

			time.Sleep(250 * time.Millisecond)

			slog.Info("Playing next song in queue")
			p.Play(0, nil)
		}()
	}
	cleanupDone.Wait()
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
