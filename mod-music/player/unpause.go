package player

import (
	"fmt"

	"github.com/gookit/slog"
)

func (p *Player) Unpause(channelID string) error {
	slog.Info("Resuming playback")

	if p.GetCurrentStatus() == StatusPlaying || p.GetCurrentStatus() == StatusError {
		return fmt.Errorf("the track is already playing (or error) %v", p.GetCurrentStatus().String())
	}

	// Set new channel if needed
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
				return fmt.Errorf("error: %v", err)
			}
		} else {
			slog.Warn("Current song is nil")

			err := p.Play(0, nil)
			if err != nil {
				return fmt.Errorf("error: %v", err)
			}
		}
		return nil
	} else {
		finished, err := p.GetStreamingSession().Finished()
		if err != nil {
			return fmt.Errorf("error: %v", err)
		}
		if finished {
			return fmt.Errorf("failed to resume audio playback: stream finished")
		}

		p.GetStreamingSession().SetPaused(false)

		if !p.GetStreamingSession().Paused() {
			p.SetCurrentStatus(StatusPlaying) // we assume it's playing which may be not 100% true
			slog.Info("Stream paused?", p.GetStreamingSession().Paused())
			slog.Warn("Audio playback", p.GetCurrentStatus().String())
			return nil
		}

		return fmt.Errorf("failed to resume audio playback: stream paused")
	}
}
