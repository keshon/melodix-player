package player

import (
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/music/history"
)

// Skip skips to the next song in the queue.
func (p *Player) Skip() {
	slog.Info("Skipping to next song")

	switch p.CurrentStatus {
	case StatusPlaying, StatusPaused:

		p.CurrentStatus = StatusResting

		if p.VoiceConnection == nil || p.CurrentSong == nil {
			return
		}

		if len(p.SkipInterrupt) == 0 {
			history := history.NewHistory()
			history.AddPlaybackCountStats(p.VoiceConnection.GuildID, p.CurrentSong.ID)

			p.SkipInterrupt <- true
			p.Play(0, nil)
		}
	case StatusResting:
		if p.CurrentSong != nil {
			if len(p.SkipInterrupt) == 0 {
				history := history.NewHistory()
				history.AddPlaybackCountStats(p.VoiceConnection.GuildID, p.CurrentSong.ID)

				p.SkipInterrupt <- true
				p.Play(0, nil)
				p.CurrentStatus = StatusPlaying
			}
		} else {
			if len(p.SkipInterrupt) == 0 {
				p.SkipInterrupt <- true
				p.Play(0, nil)
				p.CurrentStatus = StatusPlaying
			}
		}
	}

}
