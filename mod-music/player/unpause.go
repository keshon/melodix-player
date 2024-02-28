package player

import (
	"errors"
	"fmt"
	"time"

	"github.com/gookit/slog"
)

func (p *Player) Unpause(channelID string) error {
	slog.Info("Resuming playback")
	slog.Error(p.GetChannelID())

	if p.GetCurrentStatus() == StatusPlaying || p.GetCurrentStatus() == StatusError {
		return fmt.Errorf("the track is already playing (or error) %v", p.GetCurrentStatus().String())
	}

	// Switch channel if needed
	if p.GetChannelID() != channelID {
		p.SetChannelID(channelID)

		slog.Warn("Sending switch channel interrupt signal to new channel", channelID)
		p.SwitchChannelInterrupt <- true
	}

	// Check if the streaming session is initialized (start all over if not)
	if p.GetStreamingSession() == nil {
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
	} else {
		p.GetStreamingSession().SetPaused(false)

		startTime := time.Now()
		timeout := 3 * time.Second
		for time.Since(startTime) <= timeout {
			if !p.GetStreamingSession().Paused() {
				p.SetCurrentStatus(StatusPlaying)
				slog.Warn("Audio playback", p.GetCurrentStatus().String())
				return nil
			}
			time.Sleep(100 * time.Millisecond)
		}

	}

	return errors.New("failed to resume audio playback: timed out after 3 seconds")
}
