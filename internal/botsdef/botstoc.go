package botsdef

import (
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	about "github.com/keshon/melodix-player/mod-about/discord"
	helloWorld "github.com/keshon/melodix-player/mod-helloworld/discord"
	music "github.com/keshon/melodix-player/mod-music/discord"
)

var Modules = []string{"hello", "about", "music"}

// CreateBotInstance creates a new bot instance based on the module name.
//
// Parameters:
// - session: a Discord session
// - module: the name of the module ("hi" or "hello")
// Returns a Discord instance.
func CreateBotInstance(session *discordgo.Session, module string) Discord {
	switch module {
	case "hello":
		return helloWorld.NewDiscord(session)
	case "about":
		return about.NewDiscord(session)
	case "music":
		return music.NewDiscord(session)

	// ..add more cases for other modules if needed

	default:
		slog.Printf("Unknown module: %s", module)
		return nil
	}
}
