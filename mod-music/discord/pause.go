package discord

import (
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/mod-music/player"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

// handlePauseCommand handles the pause command for Discord.
func (d *Discord) handlePauseCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	if d.Player.GetCurrentStatus() != player.StatusPlaying {
		return
	}

	err := d.Player.Pause()
	if err != nil {
		slog.Error("Error pausing player:", err)
		return
	}

	embedStr := d.Player.GetCurrentStatus().StringEmoji() + " " + d.Player.GetCurrentStatus().String()
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)

	slog.Info(d.Player.GetCurrentStatus().String())
}
