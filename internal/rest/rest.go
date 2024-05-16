package rest

import (
	"io"

	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/internal/botsdef"
	"github.com/keshon/melodix-player/mods/music/history"
)

type IRest interface {
	Start(router *gin.Engine)
}

type Rest struct {
	Bots map[string]map[string]botsdef.Discord
}

func NewRest(bots map[string]map[string]botsdef.Discord) IRest {
	return &Rest{
		Bots: bots,
	}
}

func (r *Rest) Start(router *gin.Engine) {
	slog.Info("REST API routes started")

	router.GET("/", func(ctx *gin.Context) {
		toc := generateTableOfContents(router)
		ctx.JSON(http.StatusOK, gin.H{"api_methods": toc})
	})

	r.registerLogsRoutes(router.Group("/logs"))
	r.registerGuildRoutes(router.Group("/guild"))
	r.registerHistoryRoutes(router.Group("/history"))
}

type GuildInfo struct {
	GuildID string
}

type GuildSession struct {
	GuildID     string
	GuildActive bool
	BotStatus   string
}

func generateTableOfContents(router *gin.Engine) []map[string]string {
	toc := make([]map[string]string, 0, len(router.Routes()))

	for _, routeInfo := range router.Routes() {
		route := map[string]string{
			"method": routeInfo.Method,
			"path":   routeInfo.Path,
		}
		toc = append(toc, route)
	}

	return toc
}

// Examples:
// http://localhost:8080/log
// http://localhost:8080/log/download
// http://localhost:8080/log/clear
func (r *Rest) registerLogsRoutes(router *gin.RouterGroup) {
	router.GET("/", func(ctx *gin.Context) {
		file, err := os.Open("./logs/all-levels.log")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()

		b, err := io.ReadAll(file)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.Data(http.StatusOK, "text/plain", b)
	})

	router.GET("/download", func(ctx *gin.Context) {
		file, err := os.Open("./logs/all-levels.log")
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer file.Close()

		ctx.File("./logs/all-levels.log")
	})

	router.GET("/clear", func(ctx *gin.Context) {
		logFilePath := "./logs/all-levels.log"

		err := os.Truncate(logFilePath, 0)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err = slog.Flush()
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, "Log file cleared")
	})
}

// Examples:
// http://localhost:8080/guild/info/897053062030585916
// http://localhost:8080/guild/playing/897053062030585916
func (r *Rest) registerGuildRoutes(router *gin.RouterGroup) {
	router.GET("/", func(ctx *gin.Context) {
		activeSessions := []GuildInfo{}

		for guildID := range r.Bots {
			activeSessions = append(activeSessions, GuildInfo{GuildID: guildID})
		}

		ctx.JSON(http.StatusOK, activeSessions)
	})
}

// Examples:
// http://localhost:8080/history
// http://localhost:8080/history/897053062030585916
func (r *Rest) registerHistoryRoutes(router *gin.RouterGroup) {
	router.GET("/", func(ctx *gin.Context) {

		h := history.NewHistory()

		history, err := h.GetHistory("", "last_played") // You need to pass appropriate arguments for sorting
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
			return
		}

		ctx.JSON(http.StatusOK, history)
	})

	router.GET("/:guild_id", func(ctx *gin.Context) {
		guildID := ctx.Param("guild_id")

		h := history.NewHistory()

		history, err := h.GetHistory(guildID, "last_played") // You need to pass appropriate arguments for sorting
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve history"})
			return
		}

		ctx.JSON(http.StatusOK, history)
	})
}
