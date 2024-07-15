package discord

import (
	"time"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mods/music/player"
)

func (d *Discord) handleResumeCommand() {
	if d.Player.GetCurrentStatus() != player.StatusPaused && d.Player.GetCurrentStatus() != player.StatusResting {
		slog.Info("Ignoring resume command because player is not paused or resting", d.Player.GetCurrentStatus().String())
	}

	channelId, err := d.findVoiceChannelWithUser()
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

	msgText := d.Player.GetCurrentStatus().StringEmoji() + " " + d.Player.GetCurrentStatus().String()
	d.sendMessageEmbed(msgText)
}
