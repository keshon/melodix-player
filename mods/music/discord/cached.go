package discord

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/player"
)

func (d *Discord) handleCacheListCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {

	if param == "sync" {
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

		// Iterate over the files and append their names and IDs to the buffer
		for _, file := range files {
			filenameNoExt := stripExtension(file.Name())
			audioFilename := replaceSpacesWithDelimeter(filenameNoExt) + filepath.Ext(file.Name())

			// Rename the file to formatted name
			oldPath := filepath.Join(cacheGuildFolder, file.Name())
			newPath := filepath.Join(cacheGuildFolder, audioFilename)

			// Rename the file to formatted name
			err := os.Rename(oldPath, newPath)
			if err != nil {
				// Handle error if renaming fails
				fmt.Printf("Error renaming file %s to %s: %v\n", oldPath, newPath, err)
			}

			filepath := filepath.Join(cacheGuildFolder, file.Name())
			_, err = db.GetTrackByFilepath(newPath)
			if err != nil {
				db.CreateTrack(&db.Track{
					Title:    file.Name(),
					Filepath: filepath,
					Source:   player.SourceLocalFile.String(),
				})
			} else {
				db.UpdateTrack(&db.Track{
					Filepath: filepath,
					Source:   player.SourceLocalFile.String(),
				})
			}

		}

		s.ChannelMessageSend(m.ChannelID, "All cached files are synced to db.")
	}

	if param == "" {
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

}
