package player

import "github.com/gookit/slog"

// Stop stops audio playback and disconnects from the voice channel.
func (p *Player) Stop() {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	p.ClearQueue()

	if p.VoiceConnection != nil {
		err := p.VoiceConnection.Speaking(false)
		if err != nil {
			slog.Errorf("Error disconnecting voice connection: %v", err)
		}

		err = p.GetVoiceConnection().Disconnect()
		if err != nil {
			slog.Fatal(err)
		}

		p.SetVoiceConnection(nil)
	}

	if p.StreamingSession != nil {
		p.StreamingSession = nil
	}

	if p.EncodingSession != nil {
		p.EncodingSession.Stop()
		p.EncodingSession.Cleanup()
	}

	if p.CurrentSong != nil {
		p.CurrentSong = nil
	}

	p.CurrentStatus = StatusResting

}
