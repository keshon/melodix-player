package player

import (
	"time"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-discord-player/mod-music/history"
)

func (p *Player) Skip() {
	slog.Warn("Song queue length: ", len(p.GetSongQueue()))

	if p.GetVoiceConnection() == nil {
		return
	}

	if p.GetCurrentSong() == nil {
		return
	}

	h := history.NewHistory()

	if len(p.GetSongQueue()) == 0 {
		slog.Warn("..stop")
		h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID)
		p.Stop()
	} else {
		if len(p.SkipInterrupt) == 0 {
			slog.Warn("..skip to", p.songQueue[0].Title)
			h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID)
			p.SkipInterrupt <- true
			time.Sleep(250 * time.Millisecond)
			p.Play(0, nil)
		}
	}

}
