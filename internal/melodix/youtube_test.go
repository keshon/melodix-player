package melodix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSong(t *testing.T) {
	// Mock a YouTube video URL
	url := "https://www.youtube.com/watch?v=4vQ8If7f374"

	// Call NewSong to create a Song instance
	youtube := NewYoutube()
	song, err := youtube.getSpecificSongFromURL(url)

	// Check for errors
	assert.NoError(t, err, "Unexpected error during NewSong")

	// Check the properties of the Song instance
	assert.Equal(t, "Video Countdown 27 Digital   10 seconds", song.Name, "Incorrect song name")
	assert.Equal(t, url, song.UserURL, "Incorrect full URL")
	// Add more property checks based on your expectations
}

func TestNewSong_InvalidURL(t *testing.T) {
	// Mock an invalid YouTube video URL
	url := "https://www.youtube.com/watch?v=4vQ8If7f374____"

	// Call NewSong with an invalid URL
	youtube := NewYoutube()
	song, err := youtube.getSpecificSongFromURL(url)

	// Check for expected error
	assert.Error(t, err, "Expected an error for an invalid URL")
	assert.Nil(t, song, "Song should be nil for an invalid URL")
}

// Add more test cases to cover other scenarios and edge cases as needed
