package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/music/player"
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

	if d.Player.GetCurrentStatus() != player.StatusPlaying {
		// Update playlist message
		if err := updatePlaylistMessage(s, m.Message.ChannelID, pleaseWaitMessage.ID, playlist, 0); err != nil {
			slog.Warnf("Error publishing playlist: %v", err)
		}
	} else {
		// Start playing if not in enqueue-only mode
		go updatePlayingStatus(d, s, m.Message.ChannelID, pleaseWaitMessage.ID, playlist, 0)

	}
}
