package discord

import (
	"fmt"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mods/music/cache"
)

func (d *Discord) handleUploadListCommand(param string) {
	guildID := d.GuildID

	c := cache.NewCache("./upload", "./cache", guildID)

	if param == "" {

		list, err := c.ListUploadedFiles()
		listStr := ""
		for key, file := range list {
			listStr += fmt.Sprintf("` %d ` %s\n", key+1, file)
		}

		if err != nil {
			d.sendMessageEmbed(err.Error())
			return
		} else {
			d.sendMessageEmbed("📼 Uploaded files\n\nUse `" + d.prefix + "uploaded extract` to extract audio from all uploaded files to cache\n\n" + listStr)
		}

		return
	}

	if param == "extract" {
		msg := d.sendMessageEmbed("⏳ Starting to extract audio...")
		stats, err := c.ExtractAudioFromVideo()
		if err != nil {
			d.editMessageEmbed(err.Error(), msg.ID)
		} else {
			if stats == nil {
				d.editMessageEmbed("❗️ Nothing to extract", msg.ID)
				return
			}
			statsStr := "💽 Extracted audio added to cache\n\nUse `" + d.prefix + "cached` command to see available files\n\n"
			slog.Info(stats)
			for key, stat := range stats {
				statsStr += fmt.Sprintf("` %v ` %s\n", key+1, stat)
			}
			d.editMessageEmbed(statsStr, msg.ID)
		}

	} else {
		d.sendMessageEmbed("Invalid parameter. Usage: /uploaded extract")
	}
}
