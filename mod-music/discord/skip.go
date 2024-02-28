package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// handleSkipCommand handles the skip command for Discord.
func (d *Discord) handleSkipCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	embedStr := "⏩ " + "Skipping"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	skipPhrase, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Error("Error sending 'skipping' message", err)
	}

	err = d.Player.Skip()
	if err != nil {
		slog.Error("Error skipping player", err)
		return
	}

	if len(d.Player.GetSongQueue()) == 0 {
		embedStr := "⏹ " + "Stopped playback"
		embedMsg := embed.NewEmbed().
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed
		_, err = s.ChannelMessageEditEmbed(m.Message.ChannelID, skipPhrase.ID, embedMsg)
		if err != nil {
			slog.Error("Error sending 'stopped playback' message", err)
		}
	}
}
