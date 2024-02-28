// Package player provides audio playback management.
package player

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/keshon/melodix-discord-player/mod-music/pkg/dca"
)

type IPlayer interface {
	Play(startAt int, song *Song) error
	Skip() error
	Enqueue(song *Song)
	Dequeue() (*Song, error)
	ClearQueue() error
	Stop() error
	Pause() error
	Unpause() error
	Lock()
	Unlock()
	GetCurrentStatus() PlaybackStatus
	SetCurrentStatus(status PlaybackStatus)
	GetSongQueue() []*Song
	SetSongQueue(queue []*Song)
	GetVoiceConnection() *discordgo.VoiceConnection
	SetVoiceConnection(voiceConnection *discordgo.VoiceConnection)
	GetEncodingSession() *dca.EncodeSession
	GetStreamingSession() *dca.StreamingSession
	GetCurrentSong() *Song
	SetCurrentSong(song *Song)
}

type Player struct {
	sync.Mutex
	vc            *discordgo.VoiceConnection
	stream        *dca.StreamingSession
	encoding      *dca.EncodeSession
	song          *Song
	queue         []*Song
	status        PlaybackStatus
	SkipInterrupt chan bool
	StopInterrupt chan bool
}

type Song struct {
	Title       string        // Title of the song
	UserURL     string        // URL provided by the user
	DownloadURL string        // URL for downloading the song
	Thumbnail   Thumbnail     // Thumbnail image for the song
	Duration    time.Duration // Duration of the song
	ID          string        // Unique ID for the song
	Source      SongSource    // Source type of the song
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
)

func (source SongSource) String() string {
	sources := map[SongSource]string{
		SourceYouTube: "YouTube",
		SourceStream:  "Stream",
	}

	return sources[source]
}

type PlaybackStatus int32

const (
	StatusResting PlaybackStatus = iota
	StatusPlaying
	StatusPaused
	StatusError
)

func (status PlaybackStatus) String() string {
	statuses := map[PlaybackStatus]string{
		StatusResting: "Resting",
		StatusPlaying: "Playing",
		StatusPaused:  "Paused",
		StatusError:   "Error",
	}

	return statuses[status]
}

func (status PlaybackStatus) StringEmoji() string {
	statuses := map[PlaybackStatus]string{
		StatusResting: "💤",
		StatusPlaying: "▶️",
		StatusPaused:  "⏸",
	}

	return statuses[status]
}

func NewPlayer(guildID string) IPlayer {
	return &Player{
		vc:            nil,
		stream:        nil,
		encoding:      nil,
		song:          nil,
		queue:         make([]*Song, 0),
		status:        StatusResting,
		SkipInterrupt: make(chan bool, 1),
		StopInterrupt: make(chan bool, 1),
	}
}

func (p *Player) Lock() {
	p.Mutex.Lock()
}

func (p *Player) Unlock() {
	p.Mutex.Unlock()
}

func (p *Player) GetCurrentStatus() PlaybackStatus {
	return p.status
}

func (p *Player) SetCurrentStatus(status PlaybackStatus) {
	p.Lock()
	defer p.Unlock()
	p.status = status
}

func (p *Player) GetSongQueue() []*Song {
	return p.queue
}

func (p *Player) SetSongQueue(queue []*Song) {
	p.queue = queue
}

func (p *Player) GetVoiceConnection() *discordgo.VoiceConnection {
	return p.vc
}

func (p *Player) SetVoiceConnection(vc *discordgo.VoiceConnection) {
	p.Lock()
	defer p.Unlock()
	p.vc = vc
}

func (p *Player) GetCurrentSong() *Song {
	return p.song
}

func (p *Player) SetCurrentSong(song *Song) {
	p.song = song
}

func (p *Player) GetEncodingSession() *dca.EncodeSession {
	return p.encoding
}

func (p *Player) SetEncodingSession(encoding *dca.EncodeSession) {
	p.Lock()
	defer p.Unlock()
	p.encoding = encoding
}

func (p *Player) GetStreamingSession() *dca.StreamingSession {
	return p.stream
}

func (p *Player) SetStreamingSession(stream *dca.StreamingSession) {
	p.Lock()
	defer p.Unlock()
	p.stream = stream
}
