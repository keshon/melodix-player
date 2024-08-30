package player

import (
	"errors"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mods/music/media"
)

func (p *Player) Enqueue(song *media.Song) {
	slog.Info("Enqueuing:", song.Title)
	p.SetSongQueue(append(p.GetSongQueue(), song))
}

func (p *Player) Dequeue() (*media.Song, error) {
	if len(p.GetSongQueue()) == 0 {
		return nil, errors.New("queue is empty")
	}

	slog.Info("Dequeuing first track from queue:")
	for id, elem := range p.GetSongQueue() {
		slog.Warn(id, " - ", elem.Title)
	}

	firstSong := p.GetSongQueue()[0]
	p.SetSongQueue(p.GetSongQueue()[1:])

	return firstSong, nil
}

func (p *Player) ClearQueue() error {
	slog.Info("Clearing song queue")

	p.Lock()
	defer p.Unlock()

	if p.GetSongQueue() == nil {
		return nil
	}

	p.SetSongQueue(make([]*media.Song, 0))
	return nil
}
