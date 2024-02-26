package player

import "github.com/gookit/slog"

// Stop stops audio playback and disconnects from the voice channel.
func (p *Player) Stop() {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	p.StopInterrupt <- true

	p.SetCurrentStatus(StatusResting)

	if p.GetVoiceConnection() == nil {
		return
	}

	err := p.GetVoiceConnection().Disconnect()
	if err != nil {
		slog.Errorf("Error disconnecting voice connection: %v", err)
	}

	p.SetVoiceConnection(nil)
	p.SetStreamingSession(nil)
	p.SetCurrentSong(nil)

}
