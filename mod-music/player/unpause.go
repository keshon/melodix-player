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
			slog.Info("Current song is not nil")
			p.Play(0, p.GetCurrentSong())
		} else {
			slog.Warn("Current song is nil")

			err := p.Play(0, nil)
			if err != nil {
				slog.Error("Error: ", err)
			}
		}
	} else {
		slog.Info("Streaming session is not nil")
		if p.GetCurrentStatus() == StatusResting {
			// slog.Warn("call for stream cleanup")
			// p.GetStreamingSession().Stop() //!!!

			if p.GetCurrentSong() != nil {
				slog.Info("Current song is not nil")
				err := p.Play(0, p.GetCurrentSong())
				if err != nil {
					slog.Error("Error: ", err)
				}
			} else {
				slog.Warn("Current song is nil")
				err := p.Play(0, nil)
				if err != nil {
					slog.Error("Error: ", err)
				}
			}
		}
	}

	// Unpause streaming session
	p.GetStreamingSession().SetPaused(false)
	// if !p.GetStreamingSession().Paused() {
	// 	p.SetCurrentStatus(StatusPlaying)
	// }

	return nil
}
