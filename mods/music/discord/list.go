package discord

func (d *Discord) handleShowQueueCommand() {
	s := d.Session
	m := d.Message

	playlist := d.Player.GetSongQueue()
	pleaseWaitMsg := d.sendMessageEmbed("Please wait...")
	showStatusMessage(d, s, m.Message.ChannelID, pleaseWaitMsg.ID, playlist, 0, false)
}
