package melodix

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gookit/slog"
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
	song, err := client.GetVideo(url)
	if err != nil {
		return nil, err
	}

	var thumbnail youtube.Thumbnail
	if len(song.Thumbnails) > 0 {
		thumbnail = song.Thumbnails[0]
	}

	return &Song{
		Name:        song.Title,
		UserURL:     url,
		DownloadURL: song.Formats.WithAudioChannels()[0].URL,
		Duration:    song.Duration,
		Thumbnail:   thumbnail,
		ID:          song.ID,
	}, nil
}

// newSongsFromURL creates an array of Song instances using the provided YouTube URL.
func newSongsFromURL(url string) ([]*Song, error) {
	client := youtube.Client{}
	var songs []*Song

	if strings.Contains(url, "list=") {
		// It's a playlist
		playlistID := extractPlaylistID(url)
		playlistVideos, err := client.GetPlaylist(playlistID)
		if err != nil {
			return nil, err
		}

		for _, video := range playlistVideos.Videos {
			url := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.ID)
			song, err := newSongFromURL(url)
			if err != nil {
				return nil, err
			}
			songs = append(songs, song)
		}
	} else {
		// It's a single song
		song, err := newSongFromURL(url)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}

// extractPlaylistID extracts the playlist ID from the given URL.
func extractPlaylistID(url string) string {
	if strings.Contains(url, "list=") {
		splitURL := strings.Split(url, "list=")
		if len(splitURL) > 1 {
			return splitURL[1]
		}
	}
	return ""
}

// getVideoIDFromTitle retrieves the YouTube video ID from the given title.
// func getVideoIDFromTitle(title string) (string, error) {
// 	searchURL := fmt.Sprintf("https://www.youtube.com/results?search_query=%v", strings.ReplaceAll(title, " ", "+"))

// 	resp, err := http.Get(searchURL)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return "", fmt.Errorf("HTTP request failed with status code %v", resp.StatusCode)
// 	}

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	re := regexp.MustCompile(`"videoId":"([a-zA-Z0-9_-]+)"`)
// 	matches := re.FindAllStringSubmatch(string(body), -1)
// 	if len(matches) > 0 && len(matches[0]) > 1 {
// 		return "https://www.youtube.com/watch?v=" + matches[0][1], nil
// 	}

// 	return "", fmt.Errorf("No video found for the given title")
// }

// getVideoURLFromTitle retrieves the YouTube video URL from the given title.
func getVideoURLFromTitle(title string) (string, error) {
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

	re := regexp.MustCompile(`"url":"/watch\?v=([a-zA-Z0-9_-]+)(?:[^"]*?"list=([a-zA-Z0-9_-]+))?[^"]*`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) > 0 && len(matches[0]) > 1 {
		videoID := matches[0][1]
		listID := matches[0][2]
		slog.Info(matches)
		slog.Info(videoID)
		slog.Info(listID)

		url := "https://www.youtube.com/watch?v=" + videoID
		if listID != "" {
			url += "&list=" + listID
		}

		slog.Info(url)

		return url, nil
	}

	return "", fmt.Errorf("No video found for the given title")
}

// FetchSongsByID fetches songs by their IDs from the history.
func FetchSongsByID(guildID string, ids []int) ([]*Song, error) {
	h := NewHistory()
	var songs []*Song

	for _, id := range ids {
		track, err := h.GetTrackFromHistory(guildID, uint(id))
		if err != nil {
			return nil, fmt.Errorf("Error getting track from history with ID %v", id)
		}

		song, err := newSongsFromURL(track.URL)
		if err != nil {
			return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
		}

		songs = append(songs, song...)
	}

	return songs, nil
}

// FetchSongsByTitle fetches songs by their titles from YouTube.
func FetchSongsByTitle(titles []string) ([]*Song, error) {
	var songs []*Song

	for _, title := range titles {
		url, err := getVideoURLFromTitle(title)
		if err != nil {
			return nil, fmt.Errorf("Error getting YouTube video URL by title: %v", err)
		}

		songs, err = newSongsFromURL(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
		}
	}

	return songs, nil
}

// FetchSongsByURL fetches songs by their URLs.
func FetchSongsByURL(urls []string) ([]*Song, error) {
	var songs []*Song

	for _, url := range urls {
		song, err := newSongsFromURL(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
		}

		songs = append(songs, song...)
	}

	return songs, nil
}
