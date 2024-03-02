package discord

import (
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mod-music/player"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

func (d *Discord) handlePauseCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	if d.Player.GetCurrentStatus() != player.StatusPlaying {
		slog.Info("Ignoring pause command because player is not playing", d.Player.GetCurrentStatus().String())
		return
	}

	err := d.Player.Pause()
	if err != nil {
		slog.Error("Error pausing player", err)
		return
	}

	slog.Info(d.Player.GetCurrentStatus().String())

	embedStr := d.Player.GetCurrentStatus().StringEmoji() + " " + d.Player.GetCurrentStatus().String()
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Error("Error sending pause message", err)
	}

}
