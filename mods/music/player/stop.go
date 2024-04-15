package player

import (
	"fmt"

	"github.com/gookit/slog"
)

func (p *Player) Stop() error {
	slog.Info("Sending stop signal...")

	if p.GetVoiceConnection() == nil {
		return fmt.Errorf("voice connection is not initialized")
	}

	// if p.GetCurrentSong() == nil {
	// 	return fmt.Errorf("current song is missing")
	// }

	p.StopInterrupt <- true

	return nil
}
