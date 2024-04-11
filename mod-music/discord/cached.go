package discord

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func (d *Discord) handleCacheListCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Get the guild ID
	guildID := m.GuildID

	// Check if the cache folder for the guild exists
	cacheGuildFolder := filepath.Join(cacheFolder, guildID)
	_, err := os.Stat(cacheGuildFolder)
	if os.IsNotExist(err) {
		// Cache folder does not exist, send a message indicating no files are cached
		s.ChannelMessageSend(m.ChannelID, "No files cached for this guild ("+m.GuildID+").")
		return
	}

	// Get a list of files in the cache folder
	files, err := os.ReadDir(cacheGuildFolder)
	if err != nil {
		// Error reading the cache folder, send an error message
		s.ChannelMessageSend(m.ChannelID, "Error reading cache folder.")
		return
	}

	// Initialize a buffer to store the file list
	var fileList strings.Builder
	fileList.WriteString("Cached files:\n")

	// Iterate over the files and append their names and IDs to the buffer
	for _, file := range files {
		// Append file name and ID to the buffer
		fileList.WriteString(fmt.Sprintf("`%s`\n", file.Name()))
	}

	// Send the list of cached files as a message
	s.ChannelMessageSend(m.ChannelID, fileList.String())
}
