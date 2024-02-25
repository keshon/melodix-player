package player

import "github.com/gookit/slog"

// Stop stops audio playback and disconnects from the voice channel.
func (p *Player) Stop() {
	slog.Info("Stopping audio playback and disconnecting from voice channel")

	// Clear the song queue
	p.ClearQueue()

	// Cleanup streaming session
	if p.GetStreamingSession() != nil {
		p.SetStreamingSession(nil)
	}

	// Cleanup encoding session
	if p.GetEncodingSession() != nil {
		p.GetEncodingSession().Stop()
		p.GetEncodingSession().Cleanup()
	}

	// Reset current song
	p.SetCurrentSong(nil)

	// Disconnect from the voice channel
	if p.GetVoiceConnection() != nil {
		// Attempt to stop speaking
		err := p.GetVoiceConnection().Speaking(false)
		if err != nil {
			slog.Errorf("Error stopping speaking: %v", err)
		}

		// Disconnect from the voice channel
		err = p.GetVoiceConnection().Disconnect()
		if err != nil {
			slog.Errorf("Error disconnecting voice connection: %v", err)
		}

		// Reset the voice connection
		p.SetVoiceConnection(nil)
	}

	// Set status to resting
	p.SetCurrentStatus(StatusResting)

}
