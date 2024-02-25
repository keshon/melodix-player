package player

import (
	"testing"
)

func TestEnqueue(t *testing.T) {
	p := NewPlayer("123456")

	song := &Song{
		Title:   "Test Song",
		UserURL: "http://userurl.com",
		// add other necessary fields
	}

	p.Enqueue(song)

	queue := p.GetSongQueue()
	if len(queue) != 1 {
		t.Errorf("Expected queue length to be 1, got %d", len(queue))
	}

	if queue[0] != song {
		t.Error("Enqueued song is not in the queue")
	}
}

func TestDequeue(t *testing.T) {
	p := NewPlayer("123456")

	song := &Song{
		Title:   "Test Song",
		UserURL: "http://userurl.com",
		// add other necessary fields
	}

	p.Enqueue(song)

	dequeuedSong, err := p.Dequeue()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if dequeuedSong != song {
		t.Error("Dequeued song does not match enqueued song")
	}

	// Ensure the queue is empty after dequeue
	queue := p.GetSongQueue()
	if len(queue) != 0 {
		t.Errorf("Expected queue length to be 0, got %d", len(queue))
	}
}

func TestClearQueue(t *testing.T) {
	p := NewPlayer("123456")

	song := &Song{
		Title:   "Test Song",
		UserURL: "http://userurl.com",
		// add other necessary fields
	}

	p.Enqueue(song)

	p.ClearQueue()

	queue := p.GetSongQueue()
	if len(queue) != 0 {
		t.Errorf("Expected queue length to be 0 after ClearQueue, got %d", len(queue))
	}
}
