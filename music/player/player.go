// Package player provides audio playback management.
package player

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/keshon/melodix-discord-player/music/pkg/dca"
)

type Thumbnail struct {
	URL    string
	Width  uint
	Height uint
}

// SongSource represents the source type of the media.
type SongSource int32

const (
	SourceYouTube SongSource = iota
	SourceStream
)

// String returns the string representation of the SongSource.
func (source SongSource) String() string {
	sources := map[SongSource]string{
		SourceYouTube: "YouTube",
		SourceStream:  "Stream",
	}

	return sources[source]
}

// Song represents a media item with relevant information.
type Song struct {
	Title       string        // Title of the song
	UserURL     string        // URL provided by the user
	DownloadURL string        // URL for downloading the song
	Thumbnail   Thumbnail     // Thumbnail image for the song
	Duration    time.Duration // Duration of the song
	ID          string        // Unique ID for the song
	Source      SongSource    // Source type of the song
}

// PlaybackStatus represents the playback status of the Player.
type PlaybackStatus int32

const (
	StatusResting PlaybackStatus = iota
	StatusPlaying
	StatusPaused
	StatusError
)

// String returns the string representation of the PlaybackStatus.
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
		StatusError:   "⚠️",
	}

	return statuses[status]
}

// Player manages audio playback and song queue.
type Player struct {
	sync.Mutex
	VoiceConnection  *discordgo.VoiceConnection
	StreamingSession *dca.StreamingSession
	EncodingSession  *dca.EncodeSession
	SongQueue        []*Song
	CurrentSong      *Song
	CurrentStatus    PlaybackStatus
	SkipInterrupt    chan bool
}

// IPlayer defines the interface for managing audio playback and song queue.
type IPlayer interface {
	Play(startAt int, song *Song)
	Skip()
	Enqueue(song *Song)
	Dequeue() *Song
	ClearQueue()
	Stop()
	Pause()
	Unpause()
	GetCurrentStatus() PlaybackStatus
	SetCurrentStatus(status PlaybackStatus)
	GetSongQueue() []*Song
	GetVoiceConnection() *discordgo.VoiceConnection
	SetVoiceConnection(voiceConnection *discordgo.VoiceConnection)
	GetStreamingSession() *dca.StreamingSession
	GetCurrentSong() *Song
}

// NewPlayer creates a new Player instance.
func NewPlayer(guildID string) IPlayer {
	return &Player{
		VoiceConnection:  nil,
		SkipInterrupt:    make(chan bool, 1),
		StreamingSession: nil,
		EncodingSession:  nil,
		SongQueue:        make([]*Song, 0),
		CurrentSong:      nil,
		CurrentStatus:    StatusResting,
	}
}

// GetStatus returns the current playback status.
func (p *Player) GetCurrentStatus() PlaybackStatus {
	return p.CurrentStatus
}

// SetStatus sets the playback status.
func (p *Player) SetCurrentStatus(status PlaybackStatus) {
	p.Lock()
	defer p.Unlock()
	p.CurrentStatus = status
}

// GetSongQueue returns the song queue.
func (p *Player) GetSongQueue() []*Song {
	return p.SongQueue
}

// GetVoiceConnection returns the voice connection.
func (p *Player) GetVoiceConnection() *discordgo.VoiceConnection {
	return p.VoiceConnection
}

// SetVoiceConnection sets the voice connection.
func (p *Player) SetVoiceConnection(voiceConnection *discordgo.VoiceConnection) {
	p.Lock()
	defer p.Unlock()
	p.VoiceConnection = voiceConnection
}

// GetCurrentSong returns the current song being played.
func (p *Player) GetCurrentSong() *Song {
	return p.CurrentSong
}

// GetStreamingSession returns the current streaming session.
func (p *Player) GetStreamingSession() *dca.StreamingSession {
	return p.StreamingSession
}
