package melodix

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/kkdai/youtube/v2"
)

// Song represents a song with relevant information.
type Song struct {
	Name        string            // Name of the song
	UserURL     string            // URL provided by the user
	DownloadURL string            // URL for downloading the song
	Thumbnail   youtube.Thumbnail // Thumbnail image for the song
	Duration    time.Duration     // Duration of the song
	ID          string            // Unique ID for the song
}

// newSongFromURL creates a new Song instance using the provided YouTube URL.
func newSongFromURL(url string) (*Song, error) {
	client := youtube.Client{}
	sng, err := client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	var thumbnail youtube.Thumbnail
	if len(sng.Thumbnails) > 0 {
		thumbnail = sng.Thumbnails[0]
	}

	return &Song{
		Name:        sng.Title,
		UserURL:     url,
		DownloadURL: sng.Formats.WithAudioChannels()[0].URL,
		Duration:    sng.Duration,
		Thumbnail:   thumbnail,
		ID:          sng.ID,
	}, nil
}

// FetchSongByID fetches a song by its ID from the history.
func FetchSongByID(guildID string, id int) (*Song, error) {
	h := NewHistory()
	track, err := h.GetTrackFromHistory(guildID, uint(id))
	if err != nil {
		return nil, fmt.Errorf("Error getting track from history with ID %v", id)
	}

	song, err := newSongFromURL(track.URL)
	if err != nil {
		return nil, fmt.Errorf("Error fetching new song from URL: %v", err)
	}

	return song, nil
}

// FetchSongByTitle fetches a song by its title from YouTube.
func FetchSongByTitle(title string) (*Song, error) {
	url, err := getVideoIDFromTitle(title)
	if err != nil {
		return nil, fmt.Errorf("Error getting YouTube song page by URL: %v", err)
	}

	song, err := newSongFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("Error fetching new song from URL: %v", err)
	}

	return song, nil
}

// FetchSongByURL fetches a song by its URL.
func FetchSongByURL(url string) (*Song, error) {
	song, err := newSongFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("Error fetching new song from URL: %v", err)
	}

	return song, nil
}

// getVideoIDFromTitle retrieves the YouTube video ID from the given title.
func getVideoIDFromTitle(title string) (string, error) {
	searchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%v", strings.ReplaceAll(title, " ", "+"))

	resp, err := http.Get(searchURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP request failed with status code %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	re := regexp.MustCompile(`"videoId":"([a-zA-Z0-9_-]+)"`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) > 0 && len(matches[0]) > 1 {
		return "https://www.youtube.com/watch?v=" + matches[0][1], nil
	}

	return "", fmt.Errorf("No video found for the given title")
}
