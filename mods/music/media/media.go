package media

import "time"

type Song struct {
	Title     string        // Title of the song
	URL       string        // URL provided by the user
	Filepath  string        // Path/URL for downloading the song
	Thumbnail Thumbnail     // Thumbnail image for the song
	Duration  time.Duration // Duration of the song
	SongID    string        // Unique ID for the song
	Source    SongSource    // Source type of the song
}

type Thumbnail struct {
	URL    string
	Width  uint
	Height uint
}

type SongSource int32

const (
	SourceYouTube SongSource = iota
	SourceStream
	SourceLocalFile
)

func (source SongSource) String() string {
	sources := map[SongSource]string{
		SourceYouTube:   "YouTube",
		SourceStream:    "Stream",
		SourceLocalFile: "LocalFile",
	}

	return sources[source]
}
