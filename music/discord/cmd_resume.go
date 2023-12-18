package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

// handleResumeCommand handles the resume command for Discord.
func (d *Discord) handleResumeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	var phrase string

	if d.Player.GetCurrentSong() != nil {
		phrase = getStartPhrase()
	} else {
		phrase = getContinuePhrase()
	}

	d.Player.Unpause()

	embedStr := d.Player.GetCurrentStatus().StringEmoji() + " " + phrase
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}
