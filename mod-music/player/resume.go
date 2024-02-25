package player

import (
	"fmt"

	"github.com/gookit/slog"
)

// Unpause resumes audio playback.
func (p *Player) Unpause() error {
	slog.Info("Resuming playback")

	// Check if current song exists
	if p.GetCurrentSong() == nil {
		return fmt.Errorf("no song is currently playing")
	}

	// Check if the current status is playing
	if p.GetCurrentStatus() == StatusPlaying || p.GetCurrentStatus() == StatusError {
		return fmt.Errorf("the track is already playing (or error)")
	}

	// Check if the streaming session is initialized (start all over if not)
	if p.GetStreamingSession() == nil {
		p.Play(0, p.GetCurrentSong())
		//return fmt.Errorf("the streaming session is not initialized")
	}

	// Unpause streaming session
	p.GetStreamingSession().SetPaused(false)
	if !p.GetStreamingSession().Paused() {
		p.SetCurrentStatus(StatusPlaying)
	}

	return nil
}
