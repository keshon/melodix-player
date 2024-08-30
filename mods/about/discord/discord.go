package discord

import (
	"fmt"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
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
		Session:          session,
		IsInstanceActive: true,
		CommandPrefix:    config.DiscordCommandPrefix,
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

func (d *Discord) SetCommandPrefix(prefix string) {
	d.CommandPrefix = prefix
}
