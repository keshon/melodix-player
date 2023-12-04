package discord

import (
	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// handleResumeCommand handles the resume command for Discord.
func (d *Discord) handleResumeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	c, _ := s.State.Channel(m.Message.ChannelID)
	g, _ := s.State.Guild(c.GuildID)

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Message.Author.ID {
			if d.Player.GetVoiceConnection() == nil {
				conn, err := d.Session.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, true)
				if err != nil {
					slog.Errorf("Error connecting to voice channel: %v", err.Error())
					s.ChannelMessageSend(m.Message.ChannelID, "Error connecting to voice channel")
					return
				}
				d.Player.SetVoiceConnection(conn)
				conn.LogLevel = discordgo.LogWarning
			}
		}
	}

	embedStr := "▶️ **Play (or resume)**"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	d.Player.Unpause()
}
