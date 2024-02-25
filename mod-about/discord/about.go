package discord

import (
	"fmt"
	"log"
	"os"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/mod-helloworld/utils"
)

// handleAboutCommand is a function to handle the about command in Discord.
//
// It takes a Discord session and a Discord message as parameters and does not return anything.
func (d *Discord) handleAboutCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	var host string
	if os.Getenv("HOST") == "" {
		host = cfg.RestHostname
	} else {
		host = os.Getenv("HOST") // from docker environment
	}

	avatarURL := utils.InferProtocolByPort(host, 443) + host + "/avatar/random?" + fmt.Sprint(time.Now().UnixNano())
	slog.Info(avatarURL)

	title := fmt.Sprintf("ℹ️ %v — About", version.AppName)
	content := fmt.Sprintf("**%v**\n\n%v", version.AppFullName, version.AppDescription)

	buildDate := "unknown"
	if version.BuildDate != "" {
		buildDate = version.BuildDate
	}

	goVer := "unknown"
	if version.GoVersion != "" {
		goVer = version.GoVersion
	}

	embedMsg := embed.NewEmbed().
		SetDescription(fmt.Sprintf("**%v**\n\n%v", title, content)).
		AddField("```"+buildDate+"```", "Build date").
		AddField("```"+goVer+"```", "Go version").
		AddField("```Created by Innokentiy Sokolov```", "[Linkedin](https://www.linkedin.com/in/keshon), [GitHub](https://github.com/keshon), [Homepage](https://keshon.ru)").
		InlineAllFields().
		SetImage(avatarURL).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed

	_, err = s.ChannelMessageSendEmbed(m.ChannelID, embedMsg)
	if err != nil {
		log.Fatal("Error sending embed message", err)
	}
}
