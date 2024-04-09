package discord

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mod-music/sources"
)

func (d *Discord) handleCacheUrlCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	s.ChannelMessageSend(m.ChannelID, "Starting download...")

	yt := sources.NewYoutube()
	song, _ := yt.GetSongFromVideoURL(param)
	downloadURLToFile("test.mp3", song.DownloadURL)

	for _, format := range song.SongRaw.Formats.WithAudioChannels() {
		slog.Warn(format.URL)
		slog.Error(format.MimeType)
		slog.Error(format.AudioSampleRate)
		slog.Error(format.AudioQuality, "\n")
	}

	s.ChannelMessageSend(m.ChannelID, "Download complete!")
}

func downloadURLToFile(filepath string, url string) (err error) {

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status: %s", resp.Status)
	}

	// Writer the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
