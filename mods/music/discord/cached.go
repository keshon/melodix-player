package discord

import (
	"fmt"

	"github.com/keshon/melodix-player/mods/music/cache"
)

func (d *Discord) handleCacheListCommand(param string) {
	guildID := d.GuildID
	uploadsFolder := "./upload"
	cacheFolder := "./cache"

	c := cache.NewCache(uploadsFolder, cacheFolder, guildID)

	if param == "" {
		list, err := c.ListCachedFiles()
		listStr := ""
		for _, file := range list {
			listStr += fmt.Sprintf("```%s```", file)
		}

		if err != nil {
			d.sendMessageEmbed(err.Error())
			return
		} else {
			d.sendMessageEmbed("🗳 Cached files\n\n" + listStr)
		}
	}

	if param == "sync" {
		msg := d.sendMessageEmbed("⏳ Starting to sync cached files with DB...")
		added, updated, removed, err := c.SyncCachedDir()
		if err != nil {
			d.editMessageEmbed(err.Error(), msg.ID)
			return
		} else {
			d.editMessageEmbed("🗃 All cached files are synced successfully\n\nUse `"+d.prefix+"cached` command to see available files\n\n**Added:** "+fmt.Sprintf("%d", added)+"\n**Updated:** "+fmt.Sprintf("%d", updated)+"\n**Removed:** "+fmt.Sprintf("%d", removed), msg.ID)
		}
	}

}
