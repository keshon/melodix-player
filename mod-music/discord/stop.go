package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// handleStopCommand handles the stop command for Discord.
func (d *Discord) handleStopCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	err := d.Player.Stop()
	if err != nil {
		slog.Error("Error stopping:", err)
		return
	}

	embedStr := "⏹ " + "Stopped playback"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}
