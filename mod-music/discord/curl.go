package discord

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mod-music/sources"
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
	err = downloadURLToFile(videoFilePath, song.DownloadURL)
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
	audioFilename := sanitizeFilename(song.Title) + ".aac"
	audioFilePath := filepath.Join(cacheGuildFolder, audioFilename)
	err = ffpmegExtractAudioFromVideo(videoFilePath, audioFilePath)
	if err != nil {
		slog.Error("Error extracting audio:", err)
		return
	}

	// Remove the temporary video file
	err = os.Remove(videoFilePath)
	if err != nil {
		slog.Error("Error removing temporary video file:", err)
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

func ffpmegExtractAudioFromVideo(videoFilePath, audioFilePath string) error {
	cmd := exec.Command("ffmpeg", "-i", videoFilePath, "-vn", "-acodec", "copy", audioFilePath)
	err := cmd.Run()
	if err != nil {
		return err
	}

	slog.Infof("Audio extracted and saved to: %s\n", audioFilePath)
	return nil
}

func CreatePathIfNotExists(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		return err
	}
	return nil
}

func sanitizeFilename(filename string) string {
	regex := regexp.MustCompile(`[^\w\-.]+`)
	sanitized := regex.ReplaceAllString(filename, "_")

	sanitized = strings.TrimSpace(sanitized)

	if len(sanitized) > maxFilenameLength {
		sanitized = sanitized[:maxFilenameLength]
	}

	return sanitized
}
