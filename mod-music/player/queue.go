package player

import (
	"errors"

	"github.com/gookit/slog"
)

// Enqueue adds a song to the queue.
func (p *Player) Enqueue(song *Song) {
	slog.Infof("Enqueue: enqueuing song to queue: %v", song.Title)

	// p.Lock()
	// defer p.Unlock()

	p.SetSongQueue(append(p.GetSongQueue(), song))

	if p.currentSong == nil {
		p.SetCurrentSong(song)
	}
}

// Dequeue removes and returns the first song from the queue.
func (p *Player) Dequeue() (*Song, error) {
	slog.Infof("Dequeue: denqueuing song from queue: %v", p.GetSongQueue())

	// p.Lock()
	// defer p.Unlock()

	queue := p.GetSongQueue()
	if len(queue) == 0 {
		return nil, errors.New("queue is empty")
	}

	firstSong := queue[0]
	p.SetSongQueue(queue[1:])

	return firstSong, nil
}

// ClearQueue clears the song queue.
func (p *Player) ClearQueue() {
	slog.Info("ClearQueue: clearing song queue")

	p.Lock()
	defer p.Unlock()

	p.SetSongQueue(nil)
}
