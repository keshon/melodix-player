package player

import (
	"fmt"

	"github.com/gookit/slog"
)

// Stop stops the audio playback and disconnects from the voice channel.
func (p *Player) Stop() error {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	if p.GetVoiceConnection() == nil {
		return fmt.Errorf("voice connection is not initialized")
	}

	if p.GetCurrentSong() == nil {
		return fmt.Errorf("current song is missing")
	}

	p.StopInterrupt <- true

	return nil
}
