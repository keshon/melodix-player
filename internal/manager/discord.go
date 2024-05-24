package manager

import (
	"fmt"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
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
	customPrefix  string
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

	switch {
	case m.Message.Content == "melodix-prefix":
		gm.handleGetCustomPrefixCommand()
		return
	case strings.HasPrefix(m.Message.Content, "melodix-prefix-update"):
		param := gm.extractQuotedText(m.Message.Content, "melodix-prefix-update")
		gm.handleSetCustomPrefixCommand(param)
		return
	case m.Message.Content == "melodix-prefix-reset":
		gm.handleResetPrefixCommand()
		return
	}

	command, _, err := gm.splitCommandFromParameter(m.Message.Content, gm.getEffectiveCommandPrefix())
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
			gm.sendMessageEmbed(fmt.Sprintf("Guild must be registered first.\nUse `%vregister` command.", gm.getEffectiveCommandPrefix()))
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
	guildID := gm.GuildID

	exists, err := db.DoesGuildExist(guildID)
	if err != nil {
		slog.Errorf("Error checking if guild is registered: %v", err)
		return
	}

	if exists {
		gm.sendMessageEmbed("Guild is already registered")
		return
	}

	guild := db.Guild{ID: guildID, Name: ""}
	err = db.CreateGuild(guild)
	if err != nil {
		slog.Errorf("Error registering guild: %v", err)
		gm.sendMessageEmbed(fmt.Sprintf("Error registering guild\n`%v`", err))
		return
	}

	gm.setupBotInstance(guildID)
	gm.sendMessageEmbed(fmt.Sprintf("Guild registered successfully\nUse `%vhelp` to see all available commands", gm.getEffectiveCommandPrefix()))
}

func (gm *GuildManager) handleUnregisterCommand() {
	guildID := gm.GuildID

	exists, err := db.DoesGuildExist(guildID)
	if err != nil {
		slog.Errorf("Error checking if guild is registered: %v", err)
		return
	}

	if !exists {
		gm.sendMessageEmbed("Guild is not registered")
		return
	}

	err = db.DeleteGuild(guildID)
	if err != nil {
		slog.Errorf("Error unregistering guild: %v", err)
		gm.sendMessageEmbed(fmt.Sprintf("Error registering guild\n`%v`", err))
		return
	}

	gm.removeBotInstance(guildID)
	gm.sendMessageEmbed("Guild unregistered successfully")
}

func (gm *GuildManager) handleWhoamiCommand() {
	stats := fmt.Sprintf("\nGuild ID:\t%s\nChat ID:\t%s\nUser Name:\t%s\nUser ID:\t%s", gm.GuildID, gm.Message.ChannelID, gm.Message.Author.Username, gm.Message.Author.ID)
	slog.Warn("Who Am I details:", stats)
	gm.sendMessageEmbed(fmt.Sprintf("User info for **%v** was logged", gm.Message.Author.Username))
}

func (gm *GuildManager) handleSetCustomPrefixCommand(param string) {
	slog.Error(param)
	err := db.SetGuildPrefix(gm.GuildID, param)
	if err != nil {
		slog.Errorf("Error setting custom prefix: %v", err)
		gm.sendMessageEmbed(fmt.Sprintf("Error setting custom prefix\n`%v`", err.Error()))
	}
	gm.customPrefix = param
	gm.sendMessageEmbed(fmt.Sprintf("Current prefix now is `%v`", gm.getEffectiveCommandPrefix()))

	gm.removeBotInstance(gm.GuildID)
	gm.setupBotInstance(gm.GuildID)
	gm.sendMessageEmbed("Bot modules was reloaded successfully")
}

func (gm *GuildManager) handleResetPrefixCommand() {
	err := db.ResetGuildPrefix(gm.GuildID)
	if err != nil {
		slog.Errorf("Error reseting prefix", err)
		gm.sendMessageEmbed(fmt.Sprintf("Error reseting prefix\n`%v`", err.Error()))
	}
	gm.customPrefix = ""
	gm.sendMessageEmbed(fmt.Sprintf("Current prefix now is `%v`", gm.getEffectiveCommandPrefix()))

	gm.removeBotInstance(gm.GuildID)
	gm.setupBotInstance(gm.GuildID)
	gm.sendMessageEmbed("Bot modules was reloaded successfully")
}

func (gm *GuildManager) handleGetCustomPrefixCommand() {
	gm.sendMessageEmbed(fmt.Sprintf("Current prefix now is `%v`", gm.getEffectiveCommandPrefix()))
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
			botInstance.Start(id, gm.getEffectiveCommandPrefix())
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

func (gm *GuildManager) getEffectiveCommandPrefix() string {
	if gm.customPrefix != "" {
		return gm.customPrefix
	}
	return gm.commandPrefix
}

func (gm *GuildManager) extractQuotedText(input, command string) string {
	trimmedInput := strings.TrimPrefix(input, command)
	trimmedInput = strings.TrimSpace(trimmedInput)

	if len(trimmedInput) >= 2 && trimmedInput[0] == '"' && trimmedInput[len(trimmedInput)-1] == '"' {
		return trimmedInput[1 : len(trimmedInput)-1]
	}

	return ""
}

func (gm *GuildManager) sendMessageEmbed(embedStr string) *discordgo.Message {
	s := gm.Session
	m := gm.Message

	embedBody := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	msg, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedBody)
	if err != nil {
		slog.Error("Error sending pause message", err)
	}

	return msg
}
