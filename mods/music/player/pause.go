package player

import (
	"errors"
	"fmt"
	"time"

	"github.com/gookit/slog"
)

func (p *Player) Pause() error {
	slog.Info("Pausing audio playback")

	if p.GetCurrentSong() == nil {
		return fmt.Errorf("no song is currently playing")
	}

	if p.GetCurrentStatus() != StatusPlaying {
		return fmt.Errorf("the current status is not playing")
	}

	if p.GetStreamingSession() == nil {
		return fmt.Errorf("the streaming session is not initialized")
	}

	p.GetStreamingSession().SetPaused(true)

	startTime := time.Now()
	timeout := 3 * time.Second
	for time.Since(startTime) <= timeout {
		if p.GetStreamingSession().Paused() {
			p.SetCurrentStatus(StatusPaused)
			slog.Warn("Audio playback", p.GetCurrentStatus().String())
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	return errors.New("failed to pause audio playback: timed out after 3 seconds")
}
