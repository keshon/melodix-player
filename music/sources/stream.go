package sources

import (
	"net/url"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-discord-player/music/player"
)

// Youtube is a struct that encapsulates the YouTube functionality.
type Stream struct {
}

// NewStream creates a new instance of stream.
func NewStream() *Stream {
	return &Stream{}
}

// FetchStreamsByURLs fetches stream URLs into Song struct.
func (s *Stream) FetchStreamsByURLs(urls []string) ([]*player.Song, error) {
	var songs []*player.Song

	for _, elem := range urls {
		var song *player.Song

		u, err := url.Parse(elem)
		if err != nil {
			slog.Errorf("Error parsing URL: %v", err)
		}

		song = &player.Song{Name: u.Host, UserURL: u.String(), DownloadURL: u.String(), Thumbnail: player.Thumbnail{}, Duration: -1, ID: ""}
		songs = append(songs, song)
	}

	return songs, nil
}
