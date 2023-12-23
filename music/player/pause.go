package player

import "github.com/gookit/slog"

// Pause pauses audio playback.
func (p *Player) Pause() {
	slog.Info("Pausing audio playback")

	if p.VoiceConnection == nil {
		return
	}

	if p.StreamingSession == nil {
		return
	}

	if p.CurrentStatus == StatusPlaying {
		p.StreamingSession.SetPaused(true)
		p.CurrentStatus = StatusPaused
	}
}
