package player

import (
	"fmt"

	"github.com/gookit/slog"
)

// Unpause resumes audio playback.
func (p *Player) Unpause() error {
	slog.Info("Resuming playback")

	// Check if the current status is playing
	if p.GetCurrentStatus() == StatusPlaying || p.GetCurrentStatus() == StatusError {
		return fmt.Errorf("the track is already playing (or error)")
	}

	// Check if the streaming session is initialized (start all over if not)
	if p.GetStreamingSession() == nil {
		if p.GetCurrentSong() != nil {
			p.Play(0, p.GetCurrentSong())
		} else {
			p.Play(0, nil)
		}
	}

	// Unpause streaming session
	p.GetStreamingSession().SetPaused(false)
	if !p.GetStreamingSession().Paused() {
		p.SetCurrentStatus(StatusPlaying)
	}

	return nil
}
