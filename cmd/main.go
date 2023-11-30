package main

import (
	"app/internal/config"
	"app/internal/db"
	"app/internal/manager"
	"app/internal/melodix"
	"app/internal/version"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
)

var botInstances map[string]*melodix.BotInstance

func main() {
	slog.Configure(func(logger *slog.SugaredLogger) {
		f := logger.Formatter.(*slog.TextFormatter)
		f.EnableColor = true
		f.SetTemplate("[{{datetime}}] [{{level}}] [{{caller}}]\t{{message}} {{data}} {{extra}}\n")
		f.ColorTheme = slog.ColorTheme
	})

	h1 := handler.MustFileHandler("./logs/all-levels.log", handler.WithLogLevels(slog.AllLevels))
	slog.PushHandler(h1)

	// logger := slog.Std()

	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
		os.Exit(0)
	}

	slog.Info("Config loaded:\n" + config.String())

	if _, err := db.InitDB("./melodix.sqlite3"); err != nil {
		slog.Fatalf("Error initializing the database: %v", err)
		os.Exit(0)
	}

	dg, err := discordgo.New("Bot " + config.DiscordBotToken)
	if err != nil {
		slog.Fatalf("Error creating Discord session: %v", err)
		os.Exit(0)
	}

	botInstances = make(map[string]*melodix.BotInstance)

	guildManager := manager.NewGuildManager(dg, botInstances)
	guildManager.Start()

	guildIDs, err := getGuildsOrSetDefault()
	if err != nil {
		slog.Fatalf("Error retrieving or creating guilds: %v", err)
		os.Exit(0)
	}

	for _, guildID := range guildIDs {
		startBotInstances(dg, guildID)
	}

	if err := dg.Open(); err != nil {
		slog.Fatalf("Error opening Discord session: %v", err)
		os.Exit(0)
	}
	defer dg.Close()

	if config.RestEnabled {
		startRestServer(config.RestGinRelease)
	}

	slog.Infof("%v is now running. Press Ctrl+C to exit", version.AppName)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func getGuildsOrSetDefault() ([]string, error) {
	guildIDs, err := db.GetAllGuildIDs()
	if err != nil {
		return nil, err
	}

	if len(guildIDs) == 0 {
		guild := db.Guild{ID: "897053062030585916", Name: "default"} // TODO: default guild id is so-so
		if err := db.CreateGuild(guild); err != nil {
			return nil, err
		}
		guildIDs = append(guildIDs, guild.ID)
	}

	return guildIDs, nil
}

func startBotInstances(session *discordgo.Session, guildID string) {
	botInstances[guildID] = &melodix.BotInstance{
		Melodix: melodix.NewDiscord(session, guildID),
	}
	botInstances[guildID].Melodix.Start(guildID)
}

func startRestServer(isReleaseMode bool) {
	if isReleaseMode {
		gin.SetMode("release")
	}

	router := gin.Default()

	restAPI := melodix.NewRest(botInstances)
	restAPI.Start(router)

	go func() {
		port := "8080" // TODO: move out port number to .env file
		slog.Infof("REST API server started on port %v\n", port)
		if err := router.Run(":" + port); err != nil {
			slog.Fatalf("Error starting REST API server: %v", err)
		}
	}()
}
