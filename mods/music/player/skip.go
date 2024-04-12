package player

import (
	"errors"
	"time"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/mods/music/history"
)

func (p *Player) Skip() error {
	slog.Info("Skipping...")

	if p.GetVoiceConnection() == nil {
		return errors.New("voice connection is not initialized")
	}

	if p.GetCurrentSong() == nil {
		return errors.New("current song is missing")
	}

	h := history.NewHistory()

	if len(p.GetSongQueue()) == 0 {
		slog.Warn("is actually stopping...")
		h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().SongID)
		p.Stop()
	} else {
		if len(p.SkipInterrupt) == 0 {
			slog.Warn("is actually skipping to", p.GetSongQueue()[0].Title)
			h.AddPlaybackCountStats(p.GetVoiceConnection().GuildID, p.GetCurrentSong().SongID)
			p.SkipInterrupt <- true
			time.Sleep(250 * time.Millisecond)
			p.Play(0, nil)
		}
	}

	return nil
}
