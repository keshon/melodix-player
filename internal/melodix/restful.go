package melodix

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

// RestfulMelodix is a struct representing the restful API for Melodix.
type RestfulMelodix struct {
	BotInstances map[string]*BotInstance
}

// NewRestfulMelodix creates a new instance of RestfulMelodix.
func NewRestfulMelodix(botInstances map[string]*BotInstance) *RestfulMelodix {
	return &RestfulMelodix{
		BotInstances: botInstances,
	}
}

// Start registers the API routes using the provided gin.Engine.
func (rm *RestfulMelodix) Start(router *gin.Engine) {
	slog.Info("REST API routes started")

	router.GET("/", func(ctx *gin.Context) {
		toc := generateTableOfContents(router)
		ctx.JSON(http.StatusOK, gin.H{"api_methods": toc})
	})

	guildRoutes := router.Group("/guild")
	{
		rm.registerGuildRoutes(guildRoutes)
	}

	playerRoutes := router.Group("/player")
	{
		rm.registerPlayerRoutes(playerRoutes)
	}

	playlistRoutes := router.Group("/history")
	{
		rm.registerHistoryRoutes(playlistRoutes)
	}
}

// GuildInfo represents information about a guild.
type GuildInfo struct {
	GuildID string
}

// GuildSession represents the session information for a guild.
type GuildSession struct {
	GuildID          string
	GuildActive      bool
	BotStatus        string
	Queue            []*Song
	CurrentSong      *Song
	PlaybackPosition float64
}

// generateTableOfContents generates a table of contents for the API routes.
func generateTableOfContents(router *gin.Engine) []map[string]string {
	var toc []map[string]string

	// Iterate over all registered routes
	for _, routeInfo := range router.Routes() {
		route := map[string]string{
			"method": routeInfo.Method,
			"path":   routeInfo.Path,
		}
		toc = append(toc, route)
	}

	return toc
}

// registerGuildRoutes registers guild-related routes.
// http://127.0.0.1:8080/guild/info/897053062030585916
// http://127.0.0.1:8080/guild/playing/897053062030585916
func (rm *RestfulMelodix) registerGuildRoutes(router *gin.RouterGroup) {
	router.GET("/ids", func(ctx *gin.Context) {
		activeSessions := []GuildInfo{}

		for guildID := range rm.BotInstances {
			activeSessions = append(activeSessions, GuildInfo{GuildID: guildID})
		}

		ctx.JSON(http.StatusOK, activeSessions)
	})

	router.GET("/playing", func(ctx *gin.Context) {
		activeSessions := []GuildSession{}

		for guildID, bot := range rm.BotInstances {
			if bot.Melodix.Player.GetStreamingSession() == nil {
				continue
			}

			session := GuildSession{
				GuildID:          guildID,
				GuildActive:      bot.Melodix.InstanceActive,
				BotStatus:        bot.Melodix.Player.GetCurrentStatus().String(),
				Queue:            bot.Melodix.Player.GetSongQueue(),
				CurrentSong:      bot.Melodix.Player.GetCurrentSong(),
				PlaybackPosition: bot.Melodix.Player.GetStreamingSession().PlaybackPosition().Seconds(),
			}

			activeSessions = append(activeSessions, session)
		}

		ctx.JSON(http.StatusOK, activeSessions)
	})
}

// registerPlayerRoutes registers player-related routes.
// http://127.0.0.1:8080/player/play/897053062030585916?url=https://www.com/watch?v=ipFaubyDUT4
// http://127.0.0.1:8080/player/pause/897053062030585916
// http://127.0.0.1:8080/player/resume/897053062030585916
func (rm *RestfulMelodix) registerPlayerRoutes(router *gin.RouterGroup) {
	router.GET("/play/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")
		songURL := ctx.Query("url")

		if songURL == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Song URL not provided"})
			return
		}

		melodixInstance, exists := rm.BotInstances[guildID]
		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
			return
		}

		song, err := FetchSongByURL(songURL)
		if err != nil {
			slog.Warnf("Error fetching song by URL: %v", err)
			return
		}

		melodixInstance.Melodix.Player.Enqueue(song)
		if melodixInstance.Melodix.Player.GetCurrentStatus() != StatusPlaying {
			melodixInstance.Melodix.Player.Play(0, nil)
		}

		ctx.JSON(http.StatusOK, gin.H{"message": "Song added to the queue or started playing"})
	})

	router.GET("/pause/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")

		melodixInstance, exists := rm.BotInstances[guildID]
		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
			return
		}

		melodixInstance.Melodix.Player.Pause()

		ctx.JSON(http.StatusOK, gin.H{"message": "Playback paused"})
	})

	router.GET("/resume/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")

		melodixInstance, exists := rm.BotInstances[guildID]
		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
			return
		}

		melodixInstance.Melodix.Player.Unpause()

		ctx.JSON(http.StatusOK, gin.H{"message": "Playback resumed"})
	})
}

// registerHistoryRoutes registers history-related routes.
// http://127.0.0.1:8080/history
// http://127.0.0.1:8080/history/897053062030585916
func (rm *RestfulMelodix) registerHistoryRoutes(router *gin.RouterGroup) {
	router.GET("/", func(ctx *gin.Context) {

		h := NewHistory()

		// Retrieve history entries for the specified guild
		history, err := h.GetHistory("", "last_played") // You need to pass appropriate arguments for sorting
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
			return
		}

		// Respond with the history for the guild
		ctx.JSON(http.StatusOK, history)
	})

	router.GET("/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")

		h := NewHistory()

		// Retrieve history entries for the specified guild
		history, err := h.GetHistory(guildID, "last_played") // You need to pass appropriate arguments for sorting
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
			return
		}

		// Respond with the history for the guild
		ctx.JSON(http.StatusOK, history)
	})
}
