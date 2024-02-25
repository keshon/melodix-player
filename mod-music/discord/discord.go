package discord

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/mod-music/player"
	"github.com/keshon/melodix-discord-player/mod-music/utils"
)

// Discord represents the Melodix instance for Discord.
type Discord struct {
	Player               player.IPlayer
	Players              map[string]player.IPlayer
	Session              *discordgo.Session
	GuildID              string
	IsInstanceActive     bool
	prefix               string
	lastChangeAvatarTime time.Time
	rateLimitDuration    time.Duration
}

// NewDiscord creates a new instance of Discord.
func NewDiscord(session *discordgo.Session) *Discord {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &Discord{
		// Player:            player.NewPlayer(guildID),
		Players:           make(map[string]player.IPlayer),
		Session:           session,
		IsInstanceActive:  true,
		prefix:            config.DiscordCommandPrefix,
		rateLimitDuration: time.Minute * 10,
	}
}

// Start starts the Discord instance.
func (d *Discord) Start(guildID string) {
	slog.Infof(`Discord instance of mod-music started for guild id %v`, guildID)

	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
	d.Player = player.NewPlayer(guildID)
}

func (d *Discord) Stop() {
	d.IsInstanceActive = false
}

// Commands handles incoming Discord commands.
func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID {
		return
	}

	if !d.IsInstanceActive {
		return
	}

	command, parameter, err := parseCommand(m.Message.Content, d.prefix)
	if err != nil {
		return
	}

	commandAliases := [][]string{
		{"pause", "!", ">"},
		{"resume", "!", ">"},
		{"play", "p", ">"},
		{"exit", "stop", "e", "x"},
		{"list", "queue", "l", "q"},
		{"add", "a", "+"},
		{"skip", "next", "ff", ">>"},
		{"history", "time", "t"},
	}

	canonicalCommand := getCanonicalCommand(command, commandAliases)
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
		if parameter == "" && d.Player.GetCurrentStatus() == player.StatusPaused || d.Player.GetCurrentStatus() == player.StatusResting {
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
	case "history":
		d.handleHistoryCommand(s, m, parameter)
	default:
		// Unknown command
	}
}

// parseCommand parses the command and parameter from the Discord input based on the provided pattern.
func parseCommand(content, pattern string) (string, string, error) {
	if !strings.HasPrefix(content, pattern) {
		return "", "", fmt.Errorf("pattern not found")
	}

	content = content[len(pattern):] // Strip the pattern

	words := strings.Fields(content) // Split by whitespace, handling multiple spaces
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := strings.ToLower(words[0])
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}

// getCanonicalCommand gets the canonical command from aliases using the given alias.
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

// changeAvatar changes bot avatar with randomly picked avatar image within allowed rate limit
func (d *Discord) changeAvatar(s *discordgo.Session) {
	// Check if the rate limit duration has passed since the last execution
	if time.Since(d.lastChangeAvatarTime) < d.rateLimitDuration {
		slog.Info("Rate-limited. Skipping changeAvatar.")
		return
	}

	imgPath, err := utils.GetRandomImagePathFromPath("./assets/avatars")
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

	// Update the last execution time
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
