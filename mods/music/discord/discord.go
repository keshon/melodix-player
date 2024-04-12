package discord

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/mods/music/player"
	"github.com/keshon/melodix-player/mods/music/utils"
)

type Discord struct {
	Player               player.IPlayer
	Session              *discordgo.Session
	GuildID              string
	IsInstanceActive     bool
	prefix               string
	lastChangeAvatarTime time.Time
	rateLimitDuration    time.Duration
}

func NewDiscord(session *discordgo.Session) *Discord {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &Discord{
		Session:           session,
		IsInstanceActive:  true,
		prefix:            config.DiscordCommandPrefix,
		rateLimitDuration: time.Minute * 10,
	}
}

func (d *Discord) Start(guildID string) {
	slog.Infof(`Discord instance of mods/music started for guild id %v`, guildID)

	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
	d.Player = player.NewPlayer(guildID, d.Session)
}

func (d *Discord) Stop() {
	d.IsInstanceActive = false
}

func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID {
		return
	}

	if !d.IsInstanceActive {
		return
	}

	command, param, err := parseCommandAndParameter(m.Message.Content, d.prefix)
	if err != nil {
		return
	}

	aliases := [][]string{
		{"pause", "!"},
		{"resume", "r", "!>"},
		{"play", "p", ">"},
		{"stop", "x"},
		{"list", "queue", "l", "q"},
		{"add", "a", "+"},
		{"skip", "next", "ff", ">>"},
		{"history", "time", "t"},
		{"curl", "cu"},
		{"cached", "cl"},
		{"uploaded", "ul"},
	}

	canonical := getCanonicalCommand(command, aliases)
	if canonical == "" {
		return
	}

	slog.Infof("Received command \"%v\" (canonical \"%v\"), param \"%v\"", command, canonical, param)

	switch canonical {
	case "pause":
		d.handlePauseCommand(s, m)
	case "resume":
		d.handleResumeCommand(s, m)
	case "play":
		d.handlePlayCommand(s, m, param, false)
	case "skip":
		d.handleSkipCommand(s, m)
	case "list":
		d.handleShowQueueCommand(s, m)
	case "add":
		d.handlePlayCommand(s, m, param, true)
	case "stop":
		d.handleStopCommand(s, m)
	case "history":
		d.handleHistoryCommand(s, m, param)
	case "curl":
		d.handleCacheUrlCommand(s, m, param)
	case "cached":
		d.handleCacheListCommand(s, m, param)
	case "uploaded":
		d.handleUploadListCommand(s, m, param)
	}
}

func parseCommandAndParameter(content, pattern string) (string, string, error) {
	if !strings.HasPrefix(content, pattern) {
		return "", "", fmt.Errorf("pattern not found")
	}

	content = content[len(pattern):]

	words := strings.Fields(content)
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := strings.ToLower(words[0])
	param := ""
	if len(words) > 1 {
		param = strings.Join(words[1:], " ")
		param = strings.TrimSpace(param)
	}
	return command, param, nil
}

func getCanonicalCommand(alias string, commandAliases [][]string) string {
	for _, aliases := range commandAliases {
		for _, a := range aliases {
			if a == alias {
				return aliases[0]
			}
		}
	}
	return ""
}

func (d *Discord) changeAvatar(s *discordgo.Session) {
	if time.Since(d.lastChangeAvatarTime) < d.rateLimitDuration {
		//slog.Info("Rate-limited. Skipping changeAvatar.")
		return
	}

	imgPath, err := utils.GetWeightedRandomImagePath("./assets/avatars")
	if err != nil {
		slog.Errorf("Error getting avatar path: %v", err)
		return
	}

	avatar, err := utils.ReadFileToBase64(imgPath)
	if err != nil {
		fmt.Printf("Error preparing avatar: %v\n", err)
		return
	}

	_, err = s.UserUpdate("", avatar)
	if err != nil {
		slog.Errorf("Error setting the avatar: %v", err)
		return
	}

	d.lastChangeAvatarTime = time.Now()
}

func findUserVoiceState(userID string, voiceStates []*discordgo.VoiceState) (*discordgo.VoiceState, bool) {
	for _, vs := range voiceStates {
		if vs.UserID == userID {
			return vs, true
		}
	}
	return nil, false
}

func findVoiceChannelWithUser(d *Discord, s *discordgo.Session, m *discordgo.MessageCreate) (string, error) {
	channel, err := s.State.Channel(m.Message.ChannelID)
	if err != nil {
		return "", err
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return "", err
	}

	vs, found := findUserVoiceState(m.Message.Author.ID, guild.VoiceStates)
	if !found {
		return "", errors.New("user not found in voice channel")
	}
	return vs.ChannelID, nil
}
