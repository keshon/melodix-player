package discord

import (
	"github.com/keshon/melodix-player/mods/music/cache"
)

func (d *Discord) handleCacheUrlCommand(param string) {
	guildID := d.GuildID
	uploadsFolder := "./upload"
	cacheFolder := "./cache"

	c := cache.NewCache(uploadsFolder, cacheFolder, guildID)

	if param == "" {
		d.sendMessageEmbed("Error: No URL specified")
		return
	}

	waitMsg := d.sendMessageEmbed("Please wait...")

	resp, err := c.Curl(param)
	if err != nil {
		d.sendMessageEmbed(err.Error())
		return
	}

	d.editMessageEmbed(resp, waitMsg.ID)
}
