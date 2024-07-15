package discord

import (
	"errors"
	"fmt"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/mods/music/player"
)

type Discord struct {
	Player           player.IPlayer
	Session          *discordgo.Session
	Message          *discordgo.MessageCreate
	GuildID          string
	IsInstanceActive bool
	prefix           string
}

func NewDiscord(session *discordgo.Session) *Discord {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &Discord{
		Session:          session,
		Message:          nil,
		IsInstanceActive: true,
		prefix:           config.DiscordCommandPrefix,
	}
}

func (d *Discord) Start(guildID string, commandPrefix string) {
	slog.Infof(`Discord instance of 'music' module started for guild id %v`, guildID)

	d.GuildID = guildID
	d.Session.AddHandler(d.Commands)
	d.Player = player.NewPlayer(guildID, d.Session)
	d.prefix = commandPrefix
}

func (d *Discord) Stop() {
	d.IsInstanceActive = false
	err := d.Player.Stop()
	if err != nil {
		slog.Error("Error stopping player", err)
		return
	}
}

func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID {
		return
	}

	if !d.IsInstanceActive {
		return
	}

	d.Message = m

	command, param, err := d.splitCommandFromParameter(m.Message.Content, d.prefix)
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
		{"now", "n"},
	}

	canonical := getCanonicalCommand(command, aliases)
	if canonical == "" {
		return
	}

	slog.Infof("Received command \"%v\" (canonical \"%v\"), parameter \"%v\"", command, canonical, param)

	switch canonical {
	case "pause":
		d.handlePauseCommand()
	case "resume":
		d.handleResumeCommand()
	case "play":
		d.handlePlayCommand(param, false)
	case "skip":
		d.handleSkipCommand()
	case "list":
		d.handleShowQueueCommand()
	case "add":
		d.handlePlayCommand(param, true)
	case "stop":
		d.handleStopCommand()
	case "history":
		d.handleHistoryCommand(param)
	case "curl":
		d.handleCacheUrlCommand(param)
	case "cached":
		d.handleCacheListCommand(param)
	case "uploaded":
		d.handleUploadListCommand(param)
	case "now":
		d.handleNowPlayngCommand()
	}
}

func (d *Discord) splitCommandFromParameter(content, commandPrefix string) (string, string, error) {
	prefixLowercase := strings.ToLower(commandPrefix)
	contentLowercase := strings.ToLower(content)

	if !strings.HasPrefix(contentLowercase, prefixLowercase) {
		return "", "", nil // fmt.Errorf("pattern not found")
	}

	content = content[len(commandPrefix):]

	words := strings.Fields(content)
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := strings.ToLower(words[0])
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}

	command = strings.ToLower(command)
	return command, parameter, nil
}

func getCanonicalCommand(command string, commandAliases [][]string) string {
	lowerCommand := strings.ToLower(command)
	for _, aliases := range commandAliases {
		for _, alias := range aliases {
			if strings.ToLower(alias) == lowerCommand {
				return strings.ToLower(aliases[0])
			}
		}
	}
	return ""
}

func (d *Discord) findUserVoiceState(userID string, voiceStates []*discordgo.VoiceState) (*discordgo.VoiceState, bool) {
	for _, vs := range voiceStates {
		if vs.UserID == userID {
			return vs, true
		}
	}
	return nil, false
}

func (d *Discord) findVoiceChannelWithUser() (string, error) {
	s := d.Session
	m := d.Message
	channel, err := s.State.Channel(m.Message.ChannelID)
	if err != nil {
		return "", err
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return "", err
	}

	vs, found := d.findUserVoiceState(m.Message.Author.ID, guild.VoiceStates)
	if !found {
		return "", errors.New("user not found in voice channel")
	}
	return vs.ChannelID, nil
}

func (d *Discord) sendMessageEmbed(embedStr string) *discordgo.Message {
	s := d.Session
	m := d.Message

	embedBody := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	msg, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedBody)
	if err != nil {
		slog.Error("Error sending pause message", err)
	}

	return msg
}

func (d *Discord) editMessageEmbed(embedStr string, messageID string) *discordgo.Message {
	s := d.Session
	m := d.Message

	embedBody := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	msg, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, messageID, embedBody)
	if err != nil {
		slog.Error("Error sending 'stopped playback' message", err)
	}

	return msg
}

func (d *Discord) SetCommandPrefix(prefix string) {
	d.prefix = prefix
}
