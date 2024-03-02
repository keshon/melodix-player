package discord

import (
	"time"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mod-music/player"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

// handleResumeCommand handles the resume command for Discord.
func (d *Discord) handleResumeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	if d.Player.GetCurrentStatus() != player.StatusPaused && d.Player.GetCurrentStatus() != player.StatusResting {
		slog.Info("Ignoring resume command because player is not paused or resting", d.Player.GetCurrentStatus().String())
	}

	channelId, err := findVoiceChannelWithUser(d, s, m)
	if err != nil {
		slog.Error("Error finding voice channel with user", err)
	}

	err = d.Player.Unpause(channelId)
	if err != nil {
		slog.Error("Error resuming player", err)
		return
	}

	time.Sleep(250 * time.Millisecond)

	slog.Info(d.Player.GetCurrentStatus().String())

	embedStr := d.Player.GetCurrentStatus().StringEmoji() + " " + d.Player.GetCurrentStatus().String()
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Error("Error sending 'please wait' message", err)

	}
}
