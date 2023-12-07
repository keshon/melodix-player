package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

// handlePauseCommand handles the pause command for Discord.
func (d *Discord) handlePauseCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	if d.Player.GetCurrentSong() == nil {
		return
	}

	embedStr := "⏸ Pause"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	d.Player.Pause()
}
