package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// handleShowQueueCommand handles the show queue command for Discord.
func (d *Discord) handleShowQueueCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	playlist := d.Player.GetSongQueue()

	// Wait message
	embedStr := getPleaseWaitPhrase()
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Warnf("Error sending 'please wait' message: %v", err)
	}

	if d.Player.GetCurrentSong() != nil {
		// show list of songs with curent one loaded
		updatePlayingNowMessage(d, s, m.Message.ChannelID, pleaseWaitMessage.ID, playlist, 0, false)
	} else {
		// show just a list of song
		updateAddToQueueMessage(s, m.Message.ChannelID, pleaseWaitMessage.ID, playlist, 0)
	}
}
