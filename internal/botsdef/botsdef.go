package botsdef

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	aboutModule "github.com/keshon/melodix-player/mods/about/discord"
	musicModule "github.com/keshon/melodix-player/mods/music/discord"
)

type Discord interface {
	Start(guildID string)
	Stop()
}

var Modules = []string{"aboutModule", "musicModule"}

// CreateBotInstance creates a new bot instance based on the module name.
//
// Parameters:
// - session: a Discord session
// - module: the name of the module ("hi" or "hello")
// Returns a Discord instance.
func CreateBotInstance(session *discordgo.Session, module string) Discord {
	switch module {
	case "aboutModule":
		return aboutModule.NewDiscord(session)
	case "musicModule":
		return musicModule.NewDiscord(session)

	// ..add more cases for other modules if needed

	default:
		slog.Printf("Unknown module: %s", module)
		return nil
	}
}
