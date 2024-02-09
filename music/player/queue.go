package player

import (
	"errors"

	"github.com/gookit/slog"
)

var ErrQueueEmpty = errors.New("queue is empty")

// Enqueue adds a song to the queue.
func (p *Player) Enqueue(song *Song) {
	slog.Debugf("Enqueuing song to queue: %v", song.Title)

	p.Lock()
	defer p.Unlock()

	p.SongQueue = append(p.SongQueue, song)
}

// Dequeue removes and returns the first song from the queue.
func (p *Player) Dequeue() (*Song, error) {
	slog.Info("Dequeuing song and returning it from queue")

	p.Lock()
	defer p.Unlock()

	if len(p.SongQueue) == 0 {
		return nil, ErrQueueEmpty
	}

	firstSong := p.SongQueue[0]
	p.SongQueue = p.SongQueue[1:]

	return firstSong, nil
}

// ClearQueue clears the song queue.
func (p *Player) ClearQueue() {
	slog.Info("Clearing song queue")

	p.Lock()
	defer p.Unlock()

	p.SongQueue = make([]*Song, 0)
}
