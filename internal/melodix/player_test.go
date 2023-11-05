package melodix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Helper function to initialize a mock MelodixPlayer
func newMockPlayer() *MelodixPlayer {
	return &MelodixPlayer{
		SongQueue:     make([]*Song, 0),
		SkipInterrupt: make(chan bool, 1),
		CurrentStatus: StatusResting,
	}
}

func TestEnqueue(t *testing.T) {
	mp := newMockPlayer()

	song := &Song{Name: "Test Song"}
	mp.Enqueue(song)

	assert.Equal(t, 1, len(mp.SongQueue))
	assert.Equal(t, song, mp.SongQueue[0])
}

func TestDequeue(t *testing.T) {
	mp := newMockPlayer()

	song1 := &Song{Name: "Song 1"}
	song2 := &Song{Name: "Song 2"}

	mp.Enqueue(song1)
	mp.Enqueue(song2)

	dequeuedSong := mp.Dequeue()

	assert.Equal(t, song1, dequeuedSong)
	assert.Equal(t, 1, len(mp.SongQueue))
	assert.Equal(t, song2, mp.SongQueue[0])
}

func TestClearQueue(t *testing.T) {
	mp := newMockPlayer()

	song := &Song{Name: "Test Song"}

	mp.Enqueue(song)
	mp.ClearQueue()

	assert.Equal(t, 0, len(mp.SongQueue))
}
