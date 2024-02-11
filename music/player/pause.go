package player

import "github.com/gookit/slog"

// Pause pauses audio playback.
func (p *Player) Pause() {
	slog.Info("Pausing audio playback")

	// Check if the current status is playing
	if p.GetCurrentStatus() != StatusPlaying {
		return
	}

	// Update the status to paused
	p.SetCurrentStatus(StatusPaused)

	// Check if the streaming session is initialized
	if p.GetStreamingSession() == nil {
		return
	}

	// Pause the streaming session
	p.GetStreamingSession().SetPaused(true)
}
