package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/keshon/melodix-player/mods/music/cache"
)

const (
	uploadsFolder     = "./upload"
	cacheFolder       = "./cache"
	maxFilenameLength = 255
)

func (d *Discord) handleCacheUrlCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {

	c := cache.NewCache(uploadsFolder, cacheFolder, m.GuildID)

	if param == "" {
		s.ChannelMessageSend(m.ChannelID, "No URL specified")
		return
	}

	resp, err := c.Curl(param)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, resp)
}
