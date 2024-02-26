// Package player provides audio playback management.
package player

import (
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"

	"github.com/keshon/melodix-discord-player/mod-music/pkg/dca"
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
	}

	return statuses[status]
}

// Player manages audio playback and song queue.
type Player struct {
	sync.Mutex
	voiceConnection  *discordgo.VoiceConnection
	streamingSession *dca.StreamingSession
	encodingSession  *dca.EncodeSession
	songQueue        []*Song
	currentSong      *Song
	currentStatus    PlaybackStatus
	SkipInterrupt    chan bool
	StopInterrupt    chan bool
}

// IPlayer defines the interface for managing audio playback and song queue.
type IPlayer interface {
	Play(startAt int, song *Song)
	Skip()
	Enqueue(song *Song)
	Dequeue() (*Song, error)
	ClearQueue()
	Stop()
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

// NewPlayer creates a new Player instance.
func NewPlayer(guildID string) IPlayer {
	return &Player{
		voiceConnection:  nil,
		streamingSession: nil,
		encodingSession:  nil,
		songQueue:        make([]*Song, 0),
		currentSong:      nil,
		currentStatus:    StatusResting,
		SkipInterrupt:    make(chan bool, 1),
		StopInterrupt:    make(chan bool, 1),
	}
}

// Lock locks the player for exclusive access.
func (p *Player) Lock() {
	p.Mutex.Lock()
}

// Unlock unlocks the player.
func (p *Player) Unlock() {
	p.Mutex.Unlock()
}

// GetStatus returns the current playback status.
func (p *Player) GetCurrentStatus() PlaybackStatus {
	return p.currentStatus
}

// SetStatus sets the playback status.
func (p *Player) SetCurrentStatus(status PlaybackStatus) {
	p.Lock()
	defer p.Unlock()
	p.currentStatus = status
}

// GetSongQueue returns the song queue.
func (p *Player) GetSongQueue() []*Song {
	return p.songQueue
}

// SetSongQueue sets the SongQueue field.
func (p *Player) SetSongQueue(queue []*Song) {
	// p.Lock()
	// defer p.Unlock()
	p.songQueue = queue
}

// GetVoiceConnection returns the voice connection.
func (p *Player) GetVoiceConnection() *discordgo.VoiceConnection {
	return p.voiceConnection
}

// SetVoiceConnection sets the voice connection.
func (p *Player) SetVoiceConnection(voiceConnection *discordgo.VoiceConnection) {
	p.Lock()
	defer p.Unlock()
	p.voiceConnection = voiceConnection
}

// GetCurrentSong returns the current song being played.
func (p *Player) GetCurrentSong() *Song {
	return p.currentSong
}

// SetCurrentSong sets the current song.
func (p *Player) SetCurrentSong(song *Song) {
	p.currentSong = song
}

// GetEncodingSession returns the current encoding session.
func (p *Player) GetEncodingSession() *dca.EncodeSession {
	return p.encodingSession
}

// SetEncodingSession sets the current encoding session.
func (p *Player) SetEncodingSession(session *dca.EncodeSession) {
	p.Lock()
	defer p.Unlock()
	p.encodingSession = session
}

// GetStreamingSession returns the current streaming session.
func (p *Player) GetStreamingSession() *dca.StreamingSession {
	return p.streamingSession
}

// SetStreamingSession sets the current streaming session.
func (p *Player) SetStreamingSession(session *dca.StreamingSession) {
	p.Lock()
	defer p.Unlock()
	p.streamingSession = session
}
