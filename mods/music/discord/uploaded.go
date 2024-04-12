package discord

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/player"
)

func (d *Discord) handleUploadListCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {

	if param == "" {
		// Scan uploaded folder for video files
		files, err := os.ReadDir(uploadsFolder)
		if err != nil {
			slog.Error("Error reading uploaded folder:", err)
		}

		// Send to Discord chat list of found files
		var fileList strings.Builder
		fileList.WriteString("Uploaded files:\n")
		for _, file := range files {
			// Check if file is a video file
			if filepath.Ext(file.Name()) == ".mp4" || filepath.Ext(file.Name()) == ".mkv" || filepath.Ext(file.Name()) == ".webm" {
				fileList.WriteString(fmt.Sprintf("- %s\n", file.Name()))
			}
		}

		s.ChannelMessageSend(m.ChannelID, fileList.String())

		return
	}

	if param == "extract" {
		// Scan uploaded folder for video files
		files, err := os.ReadDir(uploadsFolder)
		if err != nil {
			slog.Error("Error reading uploaded folder:", err)
		}

		// Iterate each file
		for _, file := range files {

			// Check if file is a video file
			if filepath.Ext(file.Name()) == ".mp4" || filepath.Ext(file.Name()) == ".mkv" || filepath.Ext(file.Name()) == ".webm" || filepath.Ext(file.Name()) == ".flv" {

				// Check if cache folder for guild exists, create if not
				cacheGuildFolder := filepath.Join(cacheFolder, m.GuildID)
				CreatePathIfNotExists(cacheGuildFolder)

				// Extract audio from video
				videoFilePath := filepath.Join(uploadsFolder, file.Name())
				filenameNoExt := stripExtension(file.Name())
				audioFilename := replaceSpacesWithDelimeter(filenameNoExt) + ".mp3"
				audioFilePath := filepath.Join(cacheGuildFolder, audioFilename)
				err = ffmpegExtractAudioFromVideo(videoFilePath, audioFilePath)
				if err != nil {
					slog.Error("Error extracting audio:", err)
					continue
				}

				// Remove the temporary video file
				err = os.Remove(videoFilePath)
				if err != nil {
					slog.Error("Error removing temporary video file:", err)
				}

				// Check if cached file exists in database
				song, err := db.GetTrackByFilepath(audioFilename)
				if err == nil {
					song.Filepath = audioFilePath
					err := db.UpdateTrack(song)
					if err != nil {
						slog.Error("Error updating track in database:", err)
						continue
					}
				} else {
					newTrack := &db.Track{
						Title:    audioFilename,
						Source:   player.SourceLocalFile.String(),
						Filepath: audioFilePath,
					}
					err = db.CreateTrack(newTrack)
					if err != nil {
						slog.Error("Error creating track in database:", err)
						continue
					}
				}

				// Get the audio file size and format
				audioFileInfo, err := os.Stat(audioFilePath)
				if err != nil {
					slog.Error("Error getting audio file information:", err)
					return
				}

				audioFileSize := humanReadableSize(audioFileInfo.Size())

				// Send message with audio extraction information
				audioMessage := fmt.Sprintf("Audio extracted and saved!\nFile Size: %s\nFile Format: %s", audioFileSize, filepath.Ext(audioFilePath))
				s.ChannelMessageSend(m.ChannelID, audioMessage)

			}
		}

		return
	} else {
		s.ChannelMessageSend(m.ChannelID, "Invalid parameter. Usage: /uploaded extract")
	}
}
