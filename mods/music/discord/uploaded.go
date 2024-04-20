package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/keshon/melodix-player/mods/music/cache"
)

func (d *Discord) handleUploadListCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {

	c := cache.NewCache("./upload", "./cache", m.GuildID)

	if param == "" {

		fileList, err := c.ListUploadedFiles()

		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "Uploaded files:\n"+fileList)
		}

		return
	}

	if param == "extract" {

		resp, err := c.ExtractAudioFromVideo(param)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
		} else {
			s.ChannelMessageSend(m.ChannelID, resp)
		}

	} else {
		s.ChannelMessageSend(m.ChannelID, "Invalid parameter. Usage: /uploaded extract")
	}
}
