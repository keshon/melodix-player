package discord

import (
	"github.com/gookit/slog"
)

// handleStopCommand handles the stop command for Discord.
func (d *Discord) handleStopCommand() {
	err := d.Player.Stop()
	if err != nil {
		slog.Error("Error stopping player", err)
		return
	}

	d.sendMessageEmbed("⏹ " + "The playback has been stopped.\nThe queue is now empty.")
}
