package player

import "github.com/gookit/slog"

// Unpause resumes audio playback.
func (p *Player) Unpause() {
	slog.Info("Resuming playback")

	if p.VoiceConnection == nil {
		return
	}

	if p.StreamingSession != nil {
		if p.CurrentStatus == StatusPaused {
			p.StreamingSession.SetPaused(false)
			p.CurrentStatus = StatusPlaying
		}
	}

	if len(p.GetSongQueue()) > 0 {
		if p.CurrentStatus == StatusResting {
			p.Play(0, nil)
			p.CurrentStatus = StatusPlaying
		}
	}
}
