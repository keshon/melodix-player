package melodix

import (
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"
)

// Rest is a struct representing the restful API for Melodix.
type Rest struct {
	BotInstances map[string]*BotInstance
}

// NewRest creates a new instance of Rest.
func NewRest(botInstances map[string]*BotInstance) *Rest {
	return &Rest{
		BotInstances: botInstances,
	}
}

// Start registers the API routes using the provided gin.Engine.
func (r *Rest) Start(router *gin.Engine) {
	slog.Info("REST API routes started")

	router.GET("/", func(ctx *gin.Context) {
		toc := generateTableOfContents(router)
		ctx.JSON(http.StatusOK, gin.H{"api_methods": toc})
	})

	logRoutes := router.Group("/log")
	{
		r.registerLogRoutes(logRoutes)
	}

	guildRoutes := router.Group("/guild")
	{
		r.registerGuildRoutes(guildRoutes)
	}

	playerRoutes := router.Group("/player")
	{
		r.registerPlayerRoutes(playerRoutes)
	}

	playlistRoutes := router.Group("/history")
	{
		r.registerHistoryRoutes(playlistRoutes)
	}

	avatarRoutes := router.Group("/avatar")
	{
		r.registerAvatarRoutes(avatarRoutes)
	}
}

// GuildInfo represents inforation about a guild.
type GuildInfo struct {
	GuildID string
}

// GuildSession represents the session inforation for a guild.
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

// registerLogRoutes operates log file
// http://localhost:8080/log
// http://localhost:8080/log/download
// http://localhost:8080/log/clear
func (r *Rest) registerLogRoutes(router *gin.RouterGroup) {

	router.GET("/", func(ctx *gin.Context) {
		file, err := os.Open("./logs/all-levels.log")
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}
		defer func() {
			if err = file.Close(); err != nil {
				ctx.Status(http.StatusInternalServerError)
				ctx.Error(err)
			}
		}()

		b, err := io.ReadAll(file)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}

		// Set the Content-Type header to text/plain to indicate plain text content
		ctx.Header("Content-Type", "text/plain")

		// Write the log content to the response body
		ctx.String(http.StatusOK, string(b))
	})

	router.GET("/download", func(ctx *gin.Context) {
		file, err := os.Open("./logs/all-levels.log")
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}
		defer func() {
			if err = file.Close(); err != nil {
				ctx.Status(http.StatusInternalServerError)
				ctx.Error(err)
			}
		}()

		// Set the headers for a file download
		ctx.Header("Content-Description", "File Transfer")
		ctx.Header("Content-Disposition", "attachment; filename=all-levels.log")
		ctx.Header("Content-Type", "application/octet-stream")
		ctx.Header("Content-Transfer-Encoding", "binary")
		ctx.Header("Expires", "0")
		ctx.Header("Cache-Control", "must-revalidate")
		ctx.Header("Pragma", "public")

		// Copy the file content to the response body
		_, err = io.Copy(ctx.Writer, file)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}

		ctx.Status(http.StatusOK)
	})

	router.GET("/clear", func(ctx *gin.Context) {
		logFilePath := "./logs/all-levels.log"

		// Truncate the log file to clear its content
		err := os.Truncate(logFilePath, 0)
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}

		// Optionally, you can also flush the logger after truncating the file
		err = slog.Flush()
		if err != nil {
			ctx.Status(http.StatusInternalServerError)
			ctx.Error(err)
			return
		}

		ctx.JSON(http.StatusOK, "Log file cleared")
	})

}

// registerGuildRoutes registers guild-related routes.
// http://localhost:8080/guild/info/897053062030585916
// http://localhost:8080/guild/playing/897053062030585916
func (r *Rest) registerGuildRoutes(router *gin.RouterGroup) {
	router.GET("/ids", func(ctx *gin.Context) {
		activeSessions := []GuildInfo{}

		for guildID := range r.BotInstances {
			activeSessions = append(activeSessions, GuildInfo{GuildID: guildID})
		}

		ctx.JSON(http.StatusOK, activeSessions)
	})

	router.GET("/playing", func(ctx *gin.Context) {
		activeSessions := []GuildSession{}

		for guildID, bot := range r.BotInstances {
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
// http://localhost:8080/player/play/897053062030585916?url=https://www.com/watch?v=ipFaubyDUT4
// http://localhost:8080/player/pause/897053062030585916
// http://localhost:8080/player/resume/897053062030585916
func (r *Rest) registerPlayerRoutes(router *gin.RouterGroup) {
	router.GET("/play/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")
		songURL := ctx.Query("url")

		if songURL == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Song URL not provided"})
			return
		}

		melodixInstance, exists := r.BotInstances[guildID]
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

		melodixInstance, exists := r.BotInstances[guildID]
		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
			return
		}

		melodixInstance.Melodix.Player.Pause()

		ctx.JSON(http.StatusOK, gin.H{"message": "Playback paused"})
	})

	router.GET("/resume/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")

		melodixInstance, exists := r.BotInstances[guildID]
		if !exists {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "Guild not found"})
			return
		}

		melodixInstance.Melodix.Player.Unpause()

		ctx.JSON(http.StatusOK, gin.H{"message": "Playback resumed"})
	})
}

// registerHistoryRoutes registers history-related routes.
// http://localhost:8080/history
// http://localhost:8080/history/897053062030585916
func (r *Rest) registerHistoryRoutes(router *gin.RouterGroup) {
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

// registerAvatarRoutes registers avatar-related routes.
// http://localhost:8080/avatar
// http://localhost:8080/avatar/random
func (r *Rest) registerAvatarRoutes(router *gin.RouterGroup) {
	router.GET("/", func(ctx *gin.Context) {

		folderPath := "./assets/avatars"

		var imageList []string
		files, err := os.ReadDir(folderPath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		for _, file := range files {
			// Filter only files with certain extensions (you can modify this if needed)
			if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".png" {
				imageList = append(imageList, file.Name())
			}
		}

		ctx.JSON(http.StatusOK, imageList)
	})

	router.GET("/random", func(ctx *gin.Context) {

		folderPath := "./assets/avatars"

		var validFiles []string
		files, err := os.ReadDir(folderPath)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Filter only files with certain extensions (you can modify this if needed)
		for _, file := range files {
			if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".png" {
				validFiles = append(validFiles, file.Name())
			}
		}

		if len(validFiles) == 0 {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "no valid images found"})
			return
		}

		// Get a random index
		randomIndex := rand.Intn(len(validFiles))
		randomImage := validFiles[randomIndex]
		imagePath := filepath.Join(folderPath, randomImage)

		// Return the image file
		ctx.File(imagePath)
	})
}
