package discord

import "github.com/bwmarrin/discordgo"

func (d *Discord) handleUploadListCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	_, _ = s.ChannelMessageSend(m.ChannelID, "Not implemented yet!")
}
