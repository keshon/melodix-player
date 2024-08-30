// Package player provides audio playback management.
package player

import (
	"sync"

	"github.com/bwmarrin/discordgo"

	"github.com/keshon/melodix-player/mods/music/history"
	"github.com/keshon/melodix-player/mods/music/media"
	"github.com/keshon/melodix-player/mods/music/third_party/dca"
)

type IPlayer interface {
	Play(startAt int, song *media.Song) error
	Skip() error
	Enqueue(song *media.Song)
	Dequeue() (*media.Song, error)
	ClearQueue() error
	Stop() error
	Pause() error
	Unpause(channelID string) error
	Lock()
	Unlock()
	GetCurrentStatus() PlaybackStatus
	SetCurrentStatus(status PlaybackStatus)
	GetSongQueue() []*media.Song
	SetSongQueue(queue []*media.Song)
	GetVoiceConnection() *discordgo.VoiceConnection
	SetVoiceConnection(voiceConnection *discordgo.VoiceConnection)
	GetEncodingSession() *dca.EncodeSession
	GetStreamingSession() *dca.StreamingSession
	GetCurrentSong() *media.Song
	SetCurrentSong(song *media.Song)
	GetChannelID() string
	SetChannelID(channelID string)
	GetDiscordSession() *discordgo.Session
	SetDiscordSession(session *discordgo.Session)
	GetGuildID() string
	SetGuildID(guildID string)
}

type Player struct {
	sync.Mutex
	vc                     *discordgo.VoiceConnection
	stream                 *dca.StreamingSession
	encoding               *dca.EncodeSession
	song                   *media.Song
	queue                  []*media.Song
	status                 PlaybackStatus
	channelID              string
	guildID                string
	session                *discordgo.Session
	history                history.IHistory
	SkipInterrupt          chan bool
	StopInterrupt          chan bool
	SwitchChannelInterrupt chan bool
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

func NewPlayer(guildID string, session *discordgo.Session) IPlayer {
	return &Player{
		vc:                     nil,
		stream:                 nil,
		encoding:               nil,
		song:                   nil,
		queue:                  make([]*media.Song, 0),
		status:                 StatusResting,
		guildID:                guildID,
		session:                session,
		history:                history.NewHistory(),
		SkipInterrupt:          make(chan bool, 1),
		StopInterrupt:          make(chan bool, 1),
		SwitchChannelInterrupt: make(chan bool, 1),
	}
}

// Setters and Getters

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

func (p *Player) GetSongQueue() []*media.Song {
	return p.queue
}

func (p *Player) SetSongQueue(queue []*media.Song) {
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

func (p *Player) GetCurrentSong() *media.Song {
	return p.song
}

func (p *Player) SetCurrentSong(song *media.Song) {
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

func (p *Player) GetChannelID() string {
	return p.channelID
}

func (p *Player) SetChannelID(channelID string) {
	p.channelID = channelID
}

func (p *Player) GetDiscordSession() *discordgo.Session {
	return p.session
}

func (p *Player) SetDiscordSession(session *discordgo.Session) {
	p.session = session
}

func (p *Player) GetGuildID() string {
	return p.guildID
}

func (p *Player) SetGuildID(guildID string) {
	p.guildID = guildID
}

func (p *Player) GetHistory() history.IHistory {
	return p.history
}
