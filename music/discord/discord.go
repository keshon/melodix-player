package discord

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/music/player"
)

// BotInstance represents an instance of a Discord bot.
type BotInstance struct {
	Melodix *Discord
}

// Discord represents the Melodix instance for Discord.
type Discord struct {
	Player               player.IPlayer
	Session              *discordgo.Session
	GuildID              string
	InstanceActive       bool
	prefix               string
	lastChangeAvatarTime time.Time
	rateLimitDuration    time.Duration
}

// NewDiscord creates a new instance of Discord.
func NewDiscord(session *discordgo.Session, guildID string) *Discord {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &Discord{
		Player:            player.NewPlayer(guildID),
		Session:           session,
		InstanceActive:    true,
		prefix:            config.DiscordCommandPrefix,
		rateLimitDuration: time.Minute * 10,
	}
}

// Start starts the Discord instance.
func (d *Discord) Start(guildID string) {
	slog.Infof(`Discord instance started for guild id %v`, guildID)

	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
}

// Commands handles incoming Discord commands.
func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID {
		return
	}

	if !d.InstanceActive {
		return
	}

	command, parameter, err := ParseCommand(m.Message.Content, d.prefix)
	if err != nil {
		return
	}

	commandAliases := [][]string{
		{"pause", "!", ">"},
		{"resume", "play", ">"},
		{"play", "p", ">"},
		{"skip", "ff", ">>"},
		{"list", "queue", "l", "q"},
		{"add", "a", "+"},
		{"exit", "stop", "e", "x"},
		{"help", "h", "?"},
		{"history", "time", "t"},
		{"about", "v"},
	}

	canonicalCommand := GetCanonicalCommand(command, commandAliases)
	if canonicalCommand == "" {
		return
	}

	switch canonicalCommand {
	case "pause":
		if parameter == "" && d.Player.GetCurrentStatus() == player.StatusPlaying {
			d.handlePauseCommand(s, m)
			return
		}
		fallthrough
	case "resume":
		if parameter == "" && d.Player.GetCurrentStatus() != player.StatusPlaying {
			d.handleResumeCommand(s, m)
			return
		}
		fallthrough
	case "play":
		d.handlePlayCommand(s, m, parameter, false)
	case "skip":
		d.handleSkipCommand(s, m)
	case "list":
		d.handleShowQueueCommand(s, m)
	case "add":
		d.handlePlayCommand(s, m, parameter, true)
	case "exit":
		d.handleStopCommand(s, m)
	case "help":
		d.handleHelpCommand(s, m)
	case "history":
		d.handleHistoryCommand(s, m, parameter)
	case "about":
		d.handleAboutCommand(s, m)
	default:
		// Unknown command
	}
}
