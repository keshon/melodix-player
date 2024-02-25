package player

import (
	"testing"
	"time"
)

func TestNewPlayer(t *testing.T) {
	guildID := "123456"
	p := NewPlayer(guildID)

	if p.GetVoiceConnection() != nil {
		t.Error("Expected voice connection to be nil initially")
	}

	if len(p.GetSongQueue()) != 0 {
		t.Error("Expected song queue to be empty initially")
	}

	if p.GetCurrentStatus() != StatusResting {
		t.Error("Expected current status to be StatusResting initially")
	}

	// Add more assertions based on your expected initial state
}

func TestNewPlayerInterface(t *testing.T) {
	guildID := "123456"
	p := NewPlayer(guildID)
	var playerInterface IPlayer = p

	// Type assertion should not panic
	if _, ok := playerInterface.(*Player); !ok {
		t.Error("Type assertion failed")
	}

	// Additional checks for IPlayer methods
	// These checks ensure that the methods declared in IPlayer are implemented
	if playerInterface.GetCurrentStatus() != StatusResting {
		t.Error("Expected current status to be StatusResting")
	}

	// Add more checks for other IPlayer methods

	// Additional checks for the Player instance
	// These checks ensure that Player-specific methods are accessible
	if p.GetVoiceConnection() != nil {
		t.Error("Expected voice connection to be nil initially")
	}

	if len(p.GetSongQueue()) != 0 {
		t.Error("Expected song queue to be empty initially")
	}
}

func TestPlaybackStatusString(t *testing.T) {
	status := StatusPlaying
	result := status.String()

	if result != "Playing" {
		t.Errorf("Expected 'Playing', got '%s'", result)
	}

	emojiResult := status.StringEmoji()
	if emojiResult != "▶️" {
		t.Errorf("Expected '▶️', got '%s'", emojiResult)
	}
}

func TestNewSong(t *testing.T) {
	thumbnail := Thumbnail{URL: "http://example.com", Width: 100, Height: 100}
	duration := time.Second * 300
	song := &Song{
		Title:       "Test Song",
		UserURL:     "http://userurl.com",
		DownloadURL: "http://downloadurl.com",
		Duration:    duration,
		Thumbnail:   thumbnail,
	}

	if song.Title != "Test Song" {
		t.Error("Expected title to be 'Test Song'")
	}

	if song.UserURL != "http://userurl.com" {
		t.Error("Expected user URL to be 'http://userurl.com'")
	}

	if song.DownloadURL != "http://downloadurl.com" {
		t.Error("Expected download URL to be 'http://downloadurl.com'")
	}

	if song.Duration != duration {
		t.Error("Expected duration to be", duration)
	}

	if song.Thumbnail.URL != "http://example.com" {
		t.Error("Expected thumbnail URL to be 'http://example.com'")
	}

	if song.Thumbnail.Width != 100 {
		t.Error("Expected thumbnail width to be 100")
	}

	if song.Thumbnail.Height != 100 {
		t.Error("Expected thumbnail height to be 100")
	}

	// Add more assertions for other fields
}

func TestSongSourceString(t *testing.T) {
	source := SourceYouTube
	result := source.String()

	if result != "YouTube" {
		t.Errorf("Expected 'YouTube', got '%s'", result)
	}
}
