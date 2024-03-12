package sources

import (
	"fmt"
	"hash/crc32"
	"net/http"
	"net/url"

	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/mod-music/player"
)

type Stream struct{}

func NewStream() *Stream {
	return &Stream{}
}

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
				ID:          fmt.Sprintf("%d", hash),
				Source:      player.SourceStream,
			}
			songs = append(songs, song)
		} else {
			return nil, fmt.Errorf("Not a valid stream due to invalid content-type: %v", contentType)
		}
	}

	return songs, nil
}

func getContentType(url string) (string, error) {
	req, err := http.NewRequest(http.MethodHead, url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	conf, err := config.NewConfig()
	if err != nil {
		return "", fmt.Errorf("error loading config: %v", err)
	}

	req.Header.Set("User-Agent", conf.DcaUserAgent)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("error sending request: %v", err)
	}
	defer resp.Body.Close()

	return resp.Header.Get("Content-Type"), nil
}

func isValidStream(contentType string) bool {
	validContentTypes := []string{
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
		"video/3gpp",
		"video/mp4",
		"video/quicktime",
		"video/webm",
		"video/x-flv",
		"video/x-ms-video",
		"video/x-ms-wmv",
		"video/x-ms-asf",
	}

	for _, validType := range validContentTypes {
		if contentType == validType {
			return true
		}
	}

	return false
}
