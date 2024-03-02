package sources

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"net/url"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/mod-music/player"
)

// Stream is a struct that encapsulates the Stream functionality.
type Stream struct{}

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
			continue // Skip to the next iteration if URL parsing fails
		}

		// Use CRC32 to hash name as unique id
		hash := crc32.ChecksumIEEE([]byte(u.Host))

		// Fetch the stream and check the content type
		contentType, err := getContentType(u.String())
		if err != nil {
			slog.Errorf("Error fetching content type: %v", err)
			continue
		}

		if isValidStream(contentType) {
			song = &player.Song{
				Title:       u.Host,
				UserURL:     u.String(),
				DownloadURL: u.String(),
				Thumbnail:   player.Thumbnail{},
				Duration:    -1,
				ID:          fmt.Sprintf("%d", hash), // Convert hash to string
				Source:      player.SourceStream,
			}
			songs = append(songs, song)
		} else {
			return nil, fmt.Errorf("Not a valid stream due to invalid content-type: %v", contentType)
		}
	}

	return songs, nil
}

// getContentType fetches the content type of a given URL.
func getContentType(url string) (string, error) {
	resp, err := http.Head(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return resp.Header.Get("Content-Type"), nil
}

func isValidStream(contentType string) bool {
	// List of common audio and video content types for radio streams
	validContentTypes := []string{
		// Audio content types
		"application/flv",
		"application/vnd.ms-wpl",
		"audio/aac",
		"audio/basic",
		"audio/flac",
		"audio/mpeg",
		"audio/ogg",
		"audio/vnd.audible",
		"audio/vnd.dece.audio",
		"audio/vnd.dts",
		"audio/vnd.rn-realaudio",
		"audio/vnd.wave",
		"audio/webm",
		"audio/x-aiff",
		"audio/x-m4a",
		"audio/x-matroska",
		"audio/x-ms-wax",
		"audio/x-ms-wma",
		"audio/x-mpegurl",
		"audio/x-pn-realaudio",
		"audio/x-scpls",
		"audio/x-wav",
		// Video content types
		"video/3gpp",
		"video/mp4",
		"video/quicktime",
		"video/webm",
		"video/x-flv",
		"video/x-ms-video",
		"video/x-ms-wmv",
		"video/x-ms-asf",
	}

	// Check if the content type is in the list of valid content types
	for _, validType := range validContentTypes {
		if contentType == validType {
			return true
		}
	}

	// If the content type is not in the list, consider it not valid stream
	return false
}
