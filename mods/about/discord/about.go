package discord

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/melodix-player/internal/version"
)

func (d *Discord) handleAboutCommand() {
	s := d.Session
	m := d.Message

	title := "ℹ️ About"
	content := fmt.Sprintf("**%v** — %v", version.AppFullName, version.AppDescription)
	content = fmt.Sprintf("%v\n\nProject repository: https://github.com/keshon/melodix-player\n", content)

	buildDate := "unknown"
	if version.BuildDate != "" {
		buildDate = version.BuildDate
	}

	goVer := "unknown"
	if version.GoVersion != "" {
		goVer = version.GoVersion
	}

	imagePath := "assets/banner-about.png"
	imageBytes, err := os.Open(imagePath)
	if err != nil {
		slog.Error("Error opening image file:", err)
	}

	embedMsg := embed.NewEmbed().
		SetDescription(fmt.Sprintf("%v\n\n%v\n\n", title, content)).
		AddField("```"+buildDate+"```", "Build date").
		AddField("```"+goVer+"```", "Go version").
		AddField("```Innokentiy Sokolov```", "[Linkedin](https://www.linkedin.com/in/keshon), [GitHub](https://github.com/keshon), [Homepage](https://keshon.ru)").
		InlineAllFields().
		SetImage("attachment://" + filepath.Base(imagePath)).
		SetColor(0x9f00d4).MessageEmbed

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Embed: embedMsg,
		Files: []*discordgo.File{
			{
				Name:   filepath.Base(imagePath),
				Reader: imageBytes,
			},
		},
	})

	if err != nil {
		log.Fatal("Error sending embed message", err)
	}
}
