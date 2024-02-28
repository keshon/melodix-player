package player

import (
	"github.com/gookit/slog"
)

// Stop stops the audio playback and disconnects from the voice channel.
func (p *Player) Stop() error {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	if p.GetVoiceConnection() == nil {
		return nil
	}

	p.StopInterrupt <- true
	p.SetSongQueue(make([]*Song, 0))
	p.SetCurrentSong(nil)

	return nil
}
