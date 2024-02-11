package player

import (
	"errors"

	"github.com/gookit/slog"
)

// Enqueue adds a song to the queue.
func (p *Player) Enqueue(song *Song) {
	slog.Infof("Enqueue: Enqueuing song to queue: %v", song.Title)

	// Lock to ensure thread safety
	p.Lock()
	defer p.Unlock()

	// Append the song to the queue
	appendedQueue := append(p.GetSongQueue(), song)
	p.SetSongQueue(appendedQueue)
}

// Dequeue removes and returns the first song from the queue.
func (p *Player) Dequeue() (*Song, error) {
	slog.Info("Dequeuing song and returning it from queue")

	// Lock to ensure thread safety
	p.Lock()
	defer p.Unlock()

	var err = errors.New("queue is empty")

	// Check if the queue is nil or empty
	if len(p.GetSongQueue()) == 0 {
		slog.Info("Dequeue: Queue is empty")
		return nil, err
	}

	// Dequeue the first song
	firstSong := p.GetSongQueue()[0]
	p.SetSongQueue(p.GetSongQueue()[1:])

	return firstSong, nil
}

// ClearQueue clears the song queue.
func (p *Player) ClearQueue() {
	slog.Info("ClearQueue: Clearing song queue")

	// Lock to ensure thread safety
	p.Lock()
	defer p.Unlock()

	if len(p.GetSongQueue()) == 0 {
		return
	}

	// Reset the song queue to an empty slice
	p.SetSongQueue(make([]*Song, 0))
}
