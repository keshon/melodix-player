package discord

import (
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/mods/music/cache"
)

func (d *Discord) handleCacheUrlCommand(param string) {
	config, err := config.NewConfig()
	if err != nil {
		slog.Error("error loading config: %w", err)
	}

	if config.DiscordAdminUserID != d.Message.Author.ID {
		d.sendMessageEmbed("Only admins can use this command")
		return
	}

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
