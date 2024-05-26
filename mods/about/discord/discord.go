package discord

import (
	"fmt"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/mods/about/utils"
)

type Discord struct {
	Session              *discordgo.Session
	Message              *discordgo.MessageCreate
	GuildID              string
	IsInstanceActive     bool
	CommandPrefix        string
	LastChangeAvatarTime time.Time
	RateLimitDuration    time.Duration
}

func NewDiscord(session *discordgo.Session) *Discord {
	config := loadConfig()

	return &Discord{
		Session:           session,
		IsInstanceActive:  true,
		CommandPrefix:     config.DiscordCommandPrefix,
		RateLimitDuration: time.Minute * 10,
	}
}

func loadConfig() *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Fatal("Error loading config", err)
	}
	return cfg
}

func (d *Discord) Start(guildID string, commandPrefix string) {
	slog.Info("Discord instance of 'about' module started for guild id", guildID)
	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
	d.CommandPrefix = commandPrefix
}

func (d *Discord) Stop() {
	d.IsInstanceActive = false
}

func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID || !d.IsInstanceActive {
		return
	}

	d.Message = m

	command, param, err := parseCommand(m.Message.Content, d.CommandPrefix)
	if err != nil {
		return
	}

	switch getCanonicalCommand(command, [][]string{
		{"help", "h", "?"},
		{"about", "v"},
	}) {
	case "help":
		d.handleHelpCommand(param)
	case "about":
		d.handleAboutCommand()
	}
}

func parseCommand(input, pattern string) (string, string, error) {
	input = strings.ToLower(input)
	pattern = strings.ToLower(pattern)

	if !strings.HasPrefix(input, pattern) {
		return "", "", nil // fmt.Errorf("pattern not found")
	}

	input = input[len(pattern):]

	words := strings.Fields(input)
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := words[0]
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}

func getCanonicalCommand(alias string, commandAliases [][]string) string {
	alias = strings.ToLower(alias)
	for _, aliases := range commandAliases {
		for _, command := range aliases {
			if strings.ToLower(command) == alias {
				return strings.ToLower(aliases[0])
			}
		}
	}
	return ""
}

func (d *Discord) changeAvatar(s *discordgo.Session) {
	if time.Since(d.LastChangeAvatarTime) < d.RateLimitDuration {
		return
	}

	imgPath, err := utils.GetRandomImagePathFromPath("./assets/avatars")
	if err != nil {
		slog.Error("Error getting random image path:", err)
		return
	}

	avatar, err := utils.ReadFileToBase64(imgPath)
	if err != nil {
		slog.Error("Error reading file to base64:", err)
		return
	}

	_, err = s.UserUpdate("", avatar)
	if err != nil {
		slog.Error("Error updating user avatar:", err)
		return
	}

	d.LastChangeAvatarTime = time.Now()
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
	d.CommandPrefix = prefix
}
