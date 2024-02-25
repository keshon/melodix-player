package player

import (
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/mod-music/history"
)

// Skip skips to the next song in the queue.
func (p *Player) Skip() {
	slog.Info("Skipping to next song")

	// p.Lock()
	// defer p.Unlock()

	switch p.GetCurrentStatus() {
	case StatusPlaying, StatusPaused:
		// Set status to resting
		// p.CurrentStatus = StatusResting

		// Check if voice connection and current song are present
		if p.GetVoiceConnection() == nil || p.GetCurrentSong() == nil {
			return
		}

		// Check if SkipInterrupt channel is available
		if len(p.SkipInterrupt) == 0 {
			// Record playback count statistics
			h := history.NewHistory()
			h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID)

			// Send interrupt to skip
			p.SkipInterrupt <- true

			// Start playing the next song
			p.Play(0, nil)
		}

	case StatusResting:
		// Check if current song is present
		if p.GetCurrentSong() != nil {
			// Check if SkipInterrupt channel is available
			if len(p.SkipInterrupt) == 0 {
				// Record playback count statistics
				h := history.NewHistory()
				h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID)

				// Send interrupt to skip
				p.SkipInterrupt <- true

				// Start playing the next song
				p.Play(0, nil)
				p.SetCurrentStatus(StatusPlaying)
			}
		} else {
			// Check if SkipInterrupt channel is available
			if len(p.SkipInterrupt) == 0 {
				// Send interrupt to skip
				p.SkipInterrupt <- true

				// Start playing the next song
				p.Play(0, nil)
				p.SetCurrentStatus(StatusPlaying)
			}
		}
	}
}
