package manager

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/internal/botsdef"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/internal/db"
)

type IGuildManager interface {
	Start()
}

type GuildManager struct {
	Session       *discordgo.Session
	Message       *discordgo.MessageCreate
	GuildID       string
	Bots          map[string]map[string]botsdef.Discord
	commandPrefix string
}

func NewGuildManager(session *discordgo.Session, bots map[string]map[string]botsdef.Discord) IGuildManager {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config:", err)
	}

	return &GuildManager{
		Session:       session,
		GuildID:       "",
		Message:       nil,
		Bots:          bots,
		commandPrefix: config.DiscordCommandPrefix,
	}
}

func (gm *GuildManager) Start() {
	slog.Info("Discord instance of guild manager started")
	gm.Session.AddHandler(gm.Commands)
}

func (gm *GuildManager) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {

	gm.Session = s
	gm.Message = m
	gm.GuildID = m.GuildID

	command, _, err := gm.splitCommandFromParameter(m.Message.Content, gm.commandPrefix)
	if err != nil {
		slog.Error(err)
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
		{"help", "h", "?"},
		{"about", "v"},
		{"cached"},
		{"uploaded"},
	}

	var commandsList []string
	for _, alias := range aliases {
		commandsList = append(commandsList, alias...)
	}

	found := false
	for _, cmd := range commandsList {
		if strings.EqualFold(command, cmd) {
			found = true
			break
		}
	}

	if found {
		guildID := m.GuildID
		exists, err := db.DoesGuildExist(guildID)
		if err != nil {
			slog.Errorf("Error checking if guild is registered: %v", err)
			return
		}

		if !exists {
			gm.Session.ChannelMessageSend(m.Message.ChannelID, "Guild must be registered first.\nUse `"+gm.commandPrefix+"register` command.")
			return
		}
	}

	switch command {
	case "register":
		gm.handleRegisterCommand()
	case "unregister":
		gm.handleUnregisterCommand()
	case "whoami":
		gm.handleWhoamiCommand()
	}
}

func (gm *GuildManager) handleRegisterCommand() {
	channelID := gm.Message.ChannelID
	guildID := gm.GuildID

	exists, err := db.DoesGuildExist(guildID)
	if err != nil {
		slog.Errorf("Error checking if guild is registered: %v", err)
		return
	}

	if exists {
		gm.Session.ChannelMessageSend(channelID, "Guild is already registered")
		return
	}

	guild := db.Guild{ID: guildID, Name: ""}
	err = db.CreateGuild(guild)
	if err != nil {
		slog.Errorf("Error registering guild: %v", err)
		gm.Session.ChannelMessageSend(channelID, "Error registering guild")
		return
	}

	gm.setupBotInstance(guildID)
	gm.Session.ChannelMessageSend(channelID, "Guild registered successfully")
}

func (gm *GuildManager) handleUnregisterCommand() {
	channelID := gm.Message.ChannelID
	guildID := gm.GuildID

	exists, err := db.DoesGuildExist(guildID)
	if err != nil {
		slog.Errorf("Error checking if guild is registered: %v", err)
		return
	}

	if !exists {
		gm.Session.ChannelMessageSend(channelID, "Guild is not registered")
		return
	}

	err = db.DeleteGuild(guildID)
	if err != nil {
		slog.Errorf("Error unregistering guild: %v", err)
		gm.Session.ChannelMessageSend(channelID, "Error unregistering guild")
		return
	}

	gm.removeBotInstance(guildID)
	gm.Session.ChannelMessageSend(channelID, "Guild unregistered successfully")
}

func (gm *GuildManager) handleWhoamiCommand() {
	stats := fmt.Sprintf("\nGuild ID:\t%s\nChat ID:\t%s\nUser Name:\t%s\nUser ID:\t%s", gm.GuildID, gm.Message.ChannelID, gm.Message.Author.Username, gm.Message.Author.ID)
	slog.Warn("Who Am I details:", stats)
	gm.Session.ChannelMessageSend(gm.Message.ChannelID, "User info for **"+gm.Message.Author.Username+"** was sent to terminal")
}

func (gm *GuildManager) setupBotInstance(guildID string) {
	id := guildID
	session := gm.Session

	if _, ok := gm.Bots[id]; !ok {
		gm.Bots[id] = make(map[string]botsdef.Discord)
	}

	for _, module := range botsdef.Modules {
		botInstance := botsdef.CreateBotInstance(session, module)
		if botInstance != nil {
			gm.Bots[id][module] = botInstance
			botInstance.Start(id)
		}
	}
}

func (gm *GuildManager) removeBotInstance(guildID string) {
	bots, ok := gm.Bots[guildID]
	if !ok {
		return
	}

	// Iterate through modules and remove each bot
	for _, module := range botsdef.Modules {
		if bot, ok := bots[module]; ok {
			bot.Stop()
			delete(bots, module)
		}
	}

	delete(gm.Bots, guildID)
}

func (gm *GuildManager) splitCommandFromParameter(content, commandPrefix string) (string, string, error) {
	if !strings.HasPrefix(content, commandPrefix) {
		return "", "", fmt.Errorf("command prefix not found")
	}

	commandAndParams := content[len(commandPrefix):]

	words := strings.Fields(commandAndParams)
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
