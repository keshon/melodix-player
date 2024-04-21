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

// NewDiscord initializes a new Discord object with the given session and guild ID.
//
// Parameters:
// - session: a pointer to a discordgo.Session
// - guildID: a string representing the guild ID
// Returns a pointer to a Discord object.
func NewDiscord(session *discordgo.Session) *Discord {
	config := loadConfig()

	return &Discord{
		Session:           session,
		IsInstanceActive:  true,
		CommandPrefix:     config.DiscordCommandPrefix,
		RateLimitDuration: time.Minute * 10,
	}
}

// loadConfig loads the configuration and returns a pointer to config.Config.
//
// No parameters.
// Returns a pointer to config.Config.
func loadConfig() *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Fatal("Error loading config", err)
	}
	return cfg
}

// Start starts the Discord instance for the specified guild ID.
//
// guildID string
func (d *Discord) Start(guildID string) {
	slog.Info("Discord instance of mod-about started for guild ID", guildID)
	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
}

func (d *Discord) Stop() {
	d.IsInstanceActive = false
}

// Commands handles the incoming Discord commands.
//
// Parameters:
//
//	s: the Discord session
//	m: the incoming Discord message
func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID || !d.IsInstanceActive {
		return
	}

	d.Message = m

	command, _, err := parseCommand(m.Message.Content, d.CommandPrefix)
	if err != nil {
		return
	}

	switch getCanonicalCommand(command, [][]string{
		{"help", "h", "?"},
		{"about", "v"},
	}) {
	case "help":
		d.handleHelpCommand()
	case "about":
		d.handleAboutCommand()
	}
}

// parseCommand parses the input based on the provided pattern
//
// input: the input string to be parsed
// pattern: the pattern to match at the beginning of the input
// string: the parsed command
// string: the parsed parameter
// error: an error if the pattern is not found or no command is found
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

// getCanonicalCommand finds the canonical command for the given alias.
//
// Parameters:
// - alias: a string representing the alias to be searched for.
// - commandAliases: a 2D slice containing the list of command aliases.
// Return type: string
func getCanonicalCommand(alias string, commandAliases [][]string) string {
	for _, aliases := range commandAliases {
		for _, command := range aliases {
			if command == alias {
				return aliases[0]
			}
		}
	}
	return ""
}

// changeAvatar changes the avatar of the Discord user.
//
// It takes a session as a parameter and does not return anything.
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
