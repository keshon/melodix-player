package discord

// handleShowQueueCommand handles the show queue command for Discord.
func (d *Discord) handleShowQueueCommand() {
	s := d.Session
	m := d.Message
	d.changeAvatar()

	playlist := d.Player.GetSongQueue()

	// Wait message
	pleaseWaitMsg := d.sendMessageEmbed("Please wait...")

	showStatusMessage(d, s, m.Message.ChannelID, pleaseWaitMsg.ID, playlist, 0, false)

}
