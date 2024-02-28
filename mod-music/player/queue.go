package player

import (
	"errors"

	"github.com/gookit/slog"
)

// Enqueue adds a song to the queue.
func (p *Player) Enqueue(song *Song) {
	slog.Warn("Player")
	slog.Info("Enqueuing:", song.Title)
	p.SetSongQueue(append(p.GetSongQueue(), song))

	//p.SetCurrentSong(song)

}

// Dequeue removes and returns the first song from the queue.
func (p *Player) Dequeue() (*Song, error) {
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

// ClearQueue clears the song queue.
func (p *Player) ClearQueue() error {
	slog.Info("Clearing song queue...")

	p.Lock()
	defer p.Unlock()

	if p.GetSongQueue() == nil {
		return nil
	}

	p.SetSongQueue(make([]*Song, 0))
	return nil
}
