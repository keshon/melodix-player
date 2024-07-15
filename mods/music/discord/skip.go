package discord

import (
	"github.com/gookit/slog"
)

func (d *Discord) handleSkipCommand() {
	skipMsg := d.sendMessageEmbed("⏩ " + "Skipping")

	err := d.Player.Skip()
	if err != nil {
		slog.Error("Error skipping player", err)
		return
	}

	if len(d.Player.GetSongQueue()) == 0 {
		d.editMessageEmbed("⏹ "+"Stopped playback", skipMsg.ID)
	}
}
