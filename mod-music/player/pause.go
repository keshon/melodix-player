package player

import (
	"fmt"

	"github.com/gookit/slog"
)

// Pause pauses audio playback.
func (p *Player) Pause() error {
	slog.Info("Pausing audio playback")

	// Check if current song exists
	if p.GetCurrentSong() == nil {
		return fmt.Errorf("no song is currently playing")
	}

	// Check if the current status is playing
	if p.GetCurrentStatus() != StatusPlaying {
		return fmt.Errorf("the current status is not playing")
	}

	// Check if the streaming session is initialized
	if p.GetStreamingSession() == nil {
		return fmt.Errorf("the streaming session is not initialized")
	}

	// Pause the streaming session
	p.GetStreamingSession().SetPaused(true)

	// Update the status to paused
	if p.GetStreamingSession().Paused() {
		p.SetCurrentStatus(StatusPaused)
	}

	return nil
}
