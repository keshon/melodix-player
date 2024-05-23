package sources

import (
	"fmt"
	"io"

	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mods/music/history"
	"github.com/keshon/melodix-player/mods/music/player"

	kkdai_youtube "github.com/kkdai/youtube/v2"
)

type IYoutube interface {
	FetchOneByURL(url string) (*player.Song, error)
	FetchManyByURL(url string) ([]*player.Song, error)
	FetchManyByManyURLs(urls []string) ([]*player.Song, error)
	FetchManyByManyIDs(guildID string, ids []int) ([]*player.Song, error)
	FetchManyByTitle(title string) ([]*player.Song, error)
}

type Youtube struct {
	youtubeClient *kkdai_youtube.Client
}

func NewYoutube() IYoutube {
	return &Youtube{
		youtubeClient: &kkdai_youtube.Client{},
	}
}

func (y *Youtube) parseSongInfo(url string) (*player.Song, error) {
	song, err := y.youtubeClient.GetVideo(url)
	if err != nil {
		return nil, err
	}

	var thumbnail player.Thumbnail
	if len(song.Thumbnails) > 0 {
		thumbnail = player.Thumbnail(song.Thumbnails[0])
	}

	return &player.Song{
		Title:     song.Title,
		URL:       url,
		Filepath:  song.Formats.WithAudioChannels()[0].URL,
		Duration:  song.Duration,
		Thumbnail: thumbnail,
		SongID:    song.ID,
		Source:    player.SourceYouTube,
	}, nil
}

func (y *Youtube) parseSongOrPlaylistInfo(url string) ([]*player.Song, error) {
	var songs []*player.Song

	if strings.Contains(url, "list=") {
		// It's a playlist
		playlistID := y.extractPlaylistID(url)
		playlistVideos, err := y.youtubeClient.GetPlaylist(playlistID)
		if err != nil {
			if err.Error() == "extractPlaylistID failed: no playlist detected or invalid playlist ID" {
				// we assume it's a 'Youtube Mix Playlist' that kkdai_youtube doesn't support
				slog.Warn("Error fetching playlist, trying to fetchsongs via regex from url", url)

				tracks, err := y.getVideoURLsFromYoutubeMixPlaylist(url)
				if err != nil {
					return nil, err
				}

				// slog.Warn(tracks)

				tracks = y.removeDuplicateStr(tracks)

				playlistVideos = &kkdai_youtube.Playlist{

					Videos: make([]*kkdai_youtube.PlaylistEntry, len(tracks)),
				}

				// Populate the Videos field
				for i, track := range tracks {
					playlistVideos.Videos[i] = &kkdai_youtube.PlaylistEntry{
						ID: track,
					}
				}
			} else {
				return nil, err
			}
		}

		var wg sync.WaitGroup

		videoIndex := make(map[string]int)
		for i, video := range playlistVideos.Videos {
			videoIndex[video.ID] = i
		}

		for _, video := range playlistVideos.Videos {
			wg.Add(1)
			go func(videoID string) {
				defer wg.Done()

				videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
				song, err := y.parseSongInfo(videoURL)
				if err != nil {
					fmt.Printf("Error fetching song for video ID %s: %v\n", videoID, err)
					return
				}

				songs = append(songs, song)
			}(video.ID)
		}

		wg.Wait()

		sort.SliceStable(songs, func(i, j int) bool {
			indexI := videoIndex[songs[i].SongID]
			indexJ := videoIndex[songs[j].SongID]

			return indexI < indexJ
		})
	} else {
		// It's a single song
		song, err := y.parseSongInfo(url)
		if err != nil {
			return nil, err
		}
		songs = append(songs, song)
	}

	return songs, nil
}

func (y *Youtube) extractPlaylistID(url string) string {
	if strings.Contains(url, "list=") {
		splitURL := strings.Split(url, "list=")
		if len(splitURL) > 1 {
			return splitURL[1]
		}
	}
	return ""
}

func extractVideoIDs(videoURLs []string) []string {
	var videoIDs []string

	re := regexp.MustCompile(`watch\?v=([^&]+)&list=`)

	for _, url := range videoURLs {
		match := re.FindStringSubmatch(url)
		if len(match) >= 2 {
			videoIDs = append(videoIDs, match[1])
		}
	}

	return videoIDs
}

// -- URL --
func (y *Youtube) FetchOneByURL(url string) (*player.Song, error) {
	song, err := y.parseSongInfo(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching new songs from URL: %v", err)
	}
	return song, nil
}

func (y *Youtube) FetchManyByURL(url string) ([]*player.Song, error) {
	var songs []*player.Song

	song, err := y.parseSongOrPlaylistInfo(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching new songs from URL: %v", err)
	}

	songs = append(songs, song...)

	return songs, nil
}

func (y *Youtube) FetchManyByManyURLs(urls []string) ([]*player.Song, error) {
	var songs []*player.Song

	for _, url := range urls {
		song, err := y.parseSongOrPlaylistInfo(url)
		if err != nil {
			return nil, fmt.Errorf("error fetching new songs from URL: %v", err)
		}

		songs = append(songs, song...)
	}

	return songs, nil
}

// -- ID --
func (y *Youtube) FetchManyByManyIDs(guildID string, ids []int) ([]*player.Song, error) {
	h := history.NewHistory()
	var songs []*player.Song

	for _, id := range ids {
		track, err := h.GetTrackFromHistory(guildID, uint(id))
		if err != nil {
			return nil, fmt.Errorf("error getting track from history with ID %v", id)
		}

		song, err := y.parseSongOrPlaylistInfo(track.URL)
		if err != nil {
			return nil, fmt.Errorf("error fetching new songs from URL: %v", err)
		}

		songs = append(songs, song...)
	}

	return songs, nil
}

// -- Title --
func (y *Youtube) FetchManyByTitle(title string) ([]*player.Song, error) {
	var songs []*player.Song

	url, err := y.getVideoURLFromTitle(title)
	if err != nil {
		return nil, fmt.Errorf("error getting YouTube video URL by title: %v", err)
	}

	songs, err = y.parseSongOrPlaylistInfo(url)
	if err != nil {
		return nil, fmt.Errorf("error fetching new songs from URL: %v", err)
	}

	return songs, nil
}

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

	return "", fmt.Errorf("no video found for the given title")
}

func (y *Youtube) getVideoURLsFromYoutubeMixPlaylist(url string) ([]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request failed with status code %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	//slog.Error(body)

	bodyString := strings.ReplaceAll(string(body), `\u0026`, "&")

	re := regexp.MustCompile(`/watch\?v=([^&]+)&list=([^&]+)`)

	matches := re.FindAllStringSubmatch(bodyString, -1)

	var videoURLs []string
	for _, match := range matches {
		if len(match) >= 3 {
			videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s&list=%s", match[1], match[2])
			videoURLs = append(videoURLs, videoURL)
		}
	}

	videoURLs = extractVideoIDs(videoURLs)

	return videoURLs, nil
}

func (y *Youtube) removeDuplicateStr(strSlice []string) []string {
	allKeys := make(map[string]bool)
	list := []string{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}
