package discord

import (
	"fmt"
	"os"
	"os/exec"
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

func stripExtension(filename string) string {
	// Use filepath.Base to get the base name of the file
	basename := filepath.Base(filename)

	// Handle hidden files (names starting with a dot)
	if strings.HasPrefix(basename, ".") {
		return basename
	}

	// Use filepath.Ext to get the extension of the file
	extension := filepath.Ext(basename)

	// Handle case where there's no extension
	if extension == "" {
		return basename
	}

	// Remove the extension from the basename
	nameWithoutExtension := strings.TrimSuffix(basename, extension)

	return nameWithoutExtension
}

func replaceSpacesWithDelimeter(filename string) string {
	// Replace spaces with dots using strings.ReplaceAll
	newFilename := strings.ReplaceAll(filename, " ", "_")
	return newFilename
}

func humanReadableSize(size int64) string {
	const (
		b = 1 << (10 * iota)
		kb
		mb
		gb
		tb
		pb
	)
	if size < kb {
		return fmt.Sprintf("%d B", size)
	}
	if size < mb {
		return fmt.Sprintf("%.2f KB", float64(size)/kb)
	}
	if size < gb {
		return fmt.Sprintf("%.2f MB", float64(size)/mb)
	}
	if size < tb {
		return fmt.Sprintf("%.2f GB", float64(size)/gb)
	}
	if size < pb {
		return fmt.Sprintf("%.2f TB", float64(size)/tb)
	}
	return fmt.Sprintf("%.2f PB", float64(size)/pb)
}

func ffmpegExtractAudioFromVideo(videoFilePath, audioFilePath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-vn", "-acodec", "libmp3lame", "-b:a", "256k", audioFilePath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	fmt.Printf("Audio extracted and saved to: %s\n", audioFilePath)
	return nil
}

func CreatePathIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		return err
	}
	return nil
}
