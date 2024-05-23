package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"

	"github.com/keshon/melodix-player/internal/botsdef"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/internal/cron"
	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/internal/manager"
	"github.com/keshon/melodix-player/internal/rest"
	"github.com/keshon/melodix-player/internal/version"
)

func main() {
	initLogger()
	config := loadConfig()
	initDatabase()
	startCron()
	discordSession := createDiscordSession(config.DiscordBotToken)
	bots := startBotHandlers(discordSession)
	handleDiscordSession(discordSession)
	startRestServer(config, bots)
	slog.Infof("%v is now running. Press Ctrl+C to exit", version.AppFullName)
	waitForExitSignal()
}

func initLogger() {
	slog.Configure(func(logger *slog.SugaredLogger) {
		if textFormatter, ok := logger.Formatter.(*slog.TextFormatter); ok {
			textFormatter.EnableColor = true
			textFormatter.SetTemplate("[{{datetime}}] [{{level}}] [{{caller}}]\t{{message}} {{data}} {{extra}}\n")
			textFormatter.ColorTheme = slog.ColorTheme
		} else {
			slog.Error("Error: Text formatter is not a *slog.TextFormatter")
		}
	})

	// fileHandler, err := handler.NewRotateFile("./logs/all-levels.log", rotatefile.Every15Min, func(hconf *handler.Config) {
	// 	*hconf = handler.Config{
	// 		MaxSize:   1024 * 1024 * 1,
	// 		Compress:  true,
	// 		BackupNum: 1,
	// 		Levels:    slog.AllLevels,
	// 	}
	// })

	fileHandler, err := handler.NewFileHandler("./logs/all-levels.log", handler.WithLogLevels(slog.AllLevels))
	if err != nil {
		slog.Error("Error creating file handler:", err)
	} else {
		slog.PushHandler(fileHandler)
	}
}

func loadConfig() *config.Config {
	cfg, err := config.NewConfig()
	if err != nil {
		slog.Fatal("Error loading config", err)
		os.Exit(1)
	}
	slog.Info("Config loaded:\n" + cfg.String())
	return cfg
}

func initDatabase() {
	_, err := db.InitDB("./database.db")
	if err != nil {
		slog.Fatal("Error initializing the database", err)
		os.Exit(1)
	}
}

func startCron() {
	cron := cron.NewCron()
	cron.Start()
}

func createDiscordSession(token string) *discordgo.Session {
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatal("Error creating Discord session", err)
	}

	session.ShouldReconnectOnError = true // Not sure if this is needed
	return session
}

func startBotHandlers(session *discordgo.Session) map[string]map[string]botsdef.Discord {
	bots := make(map[string]map[string]botsdef.Discord)

	guildIDs, err := db.GetAllGuildIDs()
	if err != nil {
		log.Fatal("Error retrieving or creating guilds", err)
	}

	for _, id := range guildIDs {
		bots[id] = make(map[string]botsdef.Discord)

		prefix, err := db.GetGuildPrefix(id)
		if err != nil {
			log.Fatal("Error retrieving prefix for the guilds", err)
		}

		if len(prefix) == 0 {
			prefix = loadConfig().DiscordCommandPrefix
		}

		for _, module := range botsdef.Modules {
			botInstance := botsdef.CreateBotInstance(session, module)
			if botInstance != nil {
				bots[id][module] = botInstance
				botInstance.Start(id, prefix)
			}
		}
	}

	guildManager := manager.NewGuildManager(session, bots)
	guildManager.Start()

	return bots
}

func handleDiscordSession(discordSession *discordgo.Session) {
	if err := discordSession.Open(); err != nil {
		slog.Fatal("Error opening Discord session", err)
		os.Exit(1)
	}
	defer discordSession.Close()
}

func startRestServer(config *config.Config, bots map[string]map[string]botsdef.Discord) {
	if !config.RestEnabled {
		return
	}
	if config.RestGinRelease {
		gin.SetMode("release")
	}
	router := gin.Default()
	restAPI := rest.NewRest(bots)
	restAPI.Start(router)
	go func() {
		if len(config.RestHostname) == 0 {
			config.RestHostname = "localhost:8080"
			slog.Warn("Hostname is empty, setting to default:", config.RestHostname)
		}
		if err := router.Run(config.RestHostname); err != nil {
			slog.Fatal("Error starting REST API server:", err)
			return
		}
		slog.Info("REST API server started on", config.RestHostname)
	}()
}

func waitForExitSignal() {
	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-exitSignalChannel
}
