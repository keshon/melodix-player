package discord

import (
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mods/music/player"
)

func (d *Discord) handlePauseCommand() {

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

	statusAsEmoji := d.Player.GetCurrentStatus().StringEmoji()
	statusAsText := d.Player.GetCurrentStatus().String()
	d.sendMessageEmbed(statusAsEmoji + " " + statusAsText)
}
