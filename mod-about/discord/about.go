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

// handleAboutCommand is a function to handle the about command in Discord.
//
// It takes a Discord session and a Discord message as parameters and does not return anything.
func (d *Discord) handleAboutCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	title := "ℹ️ About"
	content := fmt.Sprintf("**%v** — %v", version.AppFullName, version.AppDescription)
	content = fmt.Sprintf("%v\n\nProject repo: https://github.com/keshon/melodix-player\n", content)

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
		SetDescription(fmt.Sprintf("**%v**\n\n%v\n\n", title, content)).
		AddField("```"+buildDate+"```", "Build date").
		AddField("```"+goVer+"```", "Go version").
		AddField("```Innokentiy Sokolov```", "[Linkedin](https://www.linkedin.com/in/keshon), [GitHub](https://github.com/keshon), [Homepage](https://keshon.ru)").
		InlineAllFields().
		SetImage("attachment://" + filepath.Base(imagePath)).
		SetColor(0x9f00d4).MessageEmbed

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: "Check out this image!",
		Embed:   embedMsg,
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
