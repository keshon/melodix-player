package discord

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/mods/music/player"
	"github.com/keshon/melodix-player/mods/music/sources"
)

const (
	uploadsFolder     = "./upload"
	cacheFolder       = "./cache"
	maxFilenameLength = 255
)

func (d *Discord) handleCacheUrlCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	if param == "" {
		s.ChannelMessageSend(m.ChannelID, "No URL specified")
		return
	}

	yt := sources.NewYoutube()
	song, err := yt.GetSongFromVideoURL(param)
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, "Starting download "+song.Title)

	// Generate unique filename
	filename := generateFilename()

	// Download the video
	videoFilePath := filepath.Join(uploadsFolder, filename+".mp4")
	err = downloadURLToFile(videoFilePath, song.Filepath)
	if err != nil {
		slog.Error("Error downloading audio:", err)
		return
	}

	// Get the file size and format information
	fileInfo, err := os.Stat(videoFilePath)
	if err != nil {
		slog.Error("Error getting file information:", err)
		return
	}
	fileSize := humanReadableSize(fileInfo.Size())
	fileFormat := filepath.Ext(videoFilePath)

	// Send message with download complete information
	message := fmt.Sprintf("Download complete!\nFile Size: %s\nFile Format: %s", fileSize, fileFormat)
	s.ChannelMessageSend(m.ChannelID, message)

	// Check if cache folder for guild exists, create if not
	cacheGuildFolder := filepath.Join(cacheFolder, m.GuildID)
	CreatePathIfNotExists(cacheGuildFolder)

	// Extract audio from video
	audioFilename := replaceSpacesWithDelimeter(song.Title) + ".mp3"
	audioFilePath := filepath.Join(cacheGuildFolder, audioFilename)
	err = ffmpegExtractAudioFromVideo(videoFilePath, audioFilePath)
	if err != nil {
		slog.Error("Error extracting audio:", err)
		return
	}

	// Remove the temporary video file
	err = os.Remove(videoFilePath)
	if err != nil {
		slog.Error("Error removing temporary video file:", err)
	}

	// Check if cached file exists in database
	existingTrack, err := db.GetTrackBySongID(song.SongID)
	if err == nil {
		existingTrack.Filepath = audioFilePath
		existingTrack.Source = player.SourceLocalFile.String()
		err := db.UpdateTrack(existingTrack)
		if err != nil {
			slog.Error("Error updating track in database:", err)
			return
		}
	} else {
		newTrack := &db.Track{
			SongID:   song.SongID,
			Title:    song.Title,
			URL:      song.URL,
			Source:   player.SourceLocalFile.String(),
			Filepath: audioFilePath,
		}
		err = db.CreateTrack(newTrack)
		if err != nil {
			slog.Error("Error creating track in database:", err)
			return
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

func generateFilename() string {
	return fmt.Sprintf("%d", time.Now().Unix())
}

func downloadURLToFile(filepath string, url string) (err error) {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
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

func replaceSpacesWithDelimeter(filename string) string {
	// Replace spaces with dots using strings.ReplaceAll
	newFilename := strings.ReplaceAll(filename, " ", "_")
	return newFilename
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
