package player

import (
	"github.com/gookit/slog"

	"github.com/keshon/melodix-discord-player/mod-music/history"
)

func (p *Player) Skip() {
	slog.Error("song queue length: ", len(p.GetSongQueue()))

	if p.GetVoiceConnection() == nil {
		return
	}

	if p.GetCurrentSong() == nil {
		return
	}

	h := history.NewHistory()

	if len(p.GetSongQueue()) == 0 {
		h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID)
		p.Stop()
	} else {
		if len(p.SkipInterrupt) == 0 {
			h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().ID)
			p.SkipInterrupt <- true
			p.Play(0, nil)
		}
	}

}
