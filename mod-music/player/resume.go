package player

import "github.com/gookit/slog"

// Unpause resumes audio playback.
func (p *Player) Unpause() {
	slog.Info("Resuming playback")

	// Check if voice connection exists
	if p.GetVoiceConnection() == nil {
		return
	}

	// Check if a streaming session is present
	if p.GetStreamingSession() != nil {
		// Unpause if currently paused
		if p.GetCurrentStatus() == StatusPaused {
			p.GetStreamingSession().SetPaused(false)
			p.SetCurrentStatus(StatusPlaying)
		}
	}

	// Check if there are songs in the queue
	if len(p.GetSongQueue()) > 0 {
		// If player is resting, start playing
		if p.GetCurrentStatus() == StatusResting {
			p.Play(0, nil)
			p.SetCurrentStatus(StatusPlaying)
		}
	}
}
