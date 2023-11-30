package melodix

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gookit/slog"
	"github.com/kkdai/youtube/v2"
)

// Youtube is a struct that encapsulates the YouTube functionality.
type Youtube struct {
	youtubeClient *youtube.Client
}

// NewYoutube creates a new instance of Youtube.
func NewYoutube() *Youtube {
	return &Youtube{
		youtubeClient: &youtube.Client{},
	}
}

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
func (y *Youtube) getSpecificSongFromURL(url string) (*Song, error) {
	song, err := y.youtubeClient.GetVideo(url)
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

// getSongsFromPlaylist creates an array of Song instances from a YouTube playlist.
func (y *Youtube) getAnySongsFromURL(url string) ([]*Song, error) {
	var songs []*Song

	if strings.Contains(url, "list=") {
		// It's a playlist
		playlistID := y.extractPlaylistID(url)
		playlistVideos, err := y.youtubeClient.GetPlaylist(playlistID)
		if err != nil {
			return nil, err
		}

		// Use a WaitGroup to wait for all goroutines to finish
		var wg sync.WaitGroup

		// Create a map to store the index of each video ID in the playlist
		videoIndex := make(map[string]int)
		for i, video := range playlistVideos.Videos {
			videoIndex[video.ID] = i
		}

		for _, video := range playlistVideos.Videos {
			wg.Add(1)
			go func(videoID string) {
				defer wg.Done()

				videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
				song, err := y.getSpecificSongFromURL(videoURL)
				if err != nil {
					// Handle the error, e.g., log it
					fmt.Printf("Error fetching song for video ID %s: %v\n", videoID, err)
					return
				}

				// Append the song to the songs slice
				songs = append(songs, song)
			}(video.ID)
		}

		// Wait for all goroutines to finish
		wg.Wait()

		// Sort the songs based on the order of playlistVideos.Videos
		sort.SliceStable(songs, func(i, j int) bool {
			// Get the index of each song's video ID in the playlist
			indexI := videoIndex[songs[i].ID]
			indexJ := videoIndex[songs[j].ID]

			// Compare the indices to determine the order
			return indexI < indexJ
		})
	} else {
		// It's a single song
		song, err := y.getSpecificSongFromURL(url)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}

// extractPlaylistID extracts the playlist ID from the given URL.
func (y *Youtube) extractPlaylistID(url string) string {
	if strings.Contains(url, "list=") {
		splitURL := strings.Split(url, "list=")
		if len(splitURL) > 1 {
			return splitURL[1]
		}
	}
	return ""
}

// getVideoURLFromTitle retrieves the YouTube video URL from the given title.
func (y *Youtube) getVideoURLFromTitle(title string) (string, error) {
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

	re := regexp.MustCompile(`"url":"/watch\?v=([a-zA-Z0-9_-]+)(?:\\u0026list=([a-zA-Z0-9_-]+))?[^"]*`)
	matches := re.FindAllStringSubmatch(string(body), -1)
	if len(matches) > 0 && len(matches[0]) > 1 {
		videoID := matches[0][1]
		listID := matches[0][2]

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
func (y *Youtube) FetchSongsByIDs(guildID string, ids []int) ([]*Song, error) {
	h := NewHistory()
	var songs []*Song

	for _, id := range ids {
		track, err := h.GetTrackFromHistory(guildID, uint(id))
		if err != nil {
			return nil, fmt.Errorf("Error getting track from history with ID %v", id)
		}

		song, err := y.getAnySongsFromURL(track.URL)
		if err != nil {
			return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
		}

		songs = append(songs, song...)
	}

	return songs, nil
}

// FetchSongsByTitle fetches songs by their titles from YouTube.
func (y *Youtube) FetchSongsByTitles(titles []string) ([]*Song, error) {
	var songs []*Song

	for _, title := range titles {
		url, err := y.getVideoURLFromTitle(title)
		if err != nil {
			return nil, fmt.Errorf("Error getting YouTube video URL by title: %v", err)
		}

		songs, err = y.getAnySongsFromURL(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
		}
	}

	return songs, nil
}

// FetchSongsByTitle fetches song by its title from YouTube. Or songs if the initial song was part of playlist
func (y *Youtube) FetchSongsByTitle(title string) ([]*Song, error) {
	var songs []*Song

	url, err := y.getVideoURLFromTitle(title)
	if err != nil {
		return nil, fmt.Errorf("Error getting YouTube video URL by title: %v", err)
	}

	songs, err = y.getAnySongsFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
	}

	return songs, nil
}

// FetchSongsByURL fetches songs by their URLs.
func (y *Youtube) FetchSongsByURLs(urls []string) ([]*Song, error) {
	var songs []*Song

	for _, url := range urls {
		song, err := y.getAnySongsFromURL(url)
		if err != nil {
			return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
		}

		songs = append(songs, song...)
	}

	return songs, nil
}

// FetchSongByURL fetches song by its URL. Or songs if the initial song was part of playlist
func (y *Youtube) FetchSongByURLs(url string) ([]*Song, error) {
	var songs []*Song

	song, err := y.getAnySongsFromURL(url)
	if err != nil {
		return nil, fmt.Errorf("Error fetching new songs from URL: %v", err)
	}

	songs = append(songs, song...)

	return songs, nil
}
