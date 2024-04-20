package discord

import (
	"github.com/bwmarrin/discordgo"
	"github.com/keshon/melodix-player/mods/music/cache"
)

func (d *Discord) handleCacheListCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {

	uploadsFolder := "./upload"
	cacheFolder := "./cache"

	c := cache.NewCache(uploadsFolder, cacheFolder, m.GuildID)

	if param == "sync" {
		err := c.SyncCachedDir()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "Cached files are synced.")
		}
	}

	if param == "" {

		fileList, err := c.ListCachedFiles()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		} else {
			s.ChannelMessageSend(m.ChannelID, "Cached files:\n"+fileList)
		}

	}

}
