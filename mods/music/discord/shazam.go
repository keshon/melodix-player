package discord

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mods/music/player"
)

// handleShazamCommand checks the status and source of the player and fetches metadata if conditions are met
func (d *Discord) handleShazamCommand() {
	if d.Player.GetCurrentStatus() != player.StatusPlaying {
		d.sendMessageEmbed("Player must be playing to shazam")
		return // Return early if player is not playing
	}

	if d.Player.GetCurrentSong().Source != player.SourceStream {
		d.sendMessageEmbed("Only streams can be shazamed")
		return // Return early if current song source is not a stream
	}

	url := d.Player.GetCurrentSong().Filepath
	titleArtist, streamURL, err := fetchMetadata(url)
	if err != nil {
		d.sendMessageEmbed(err.Error())
		return
	}
	d.sendMessageEmbed(fmt.Sprintf("[%s](%s)", titleArtist, streamURL))
}

// fetchMetadata runs the ffmpeg command to get metadata from the stream URL
func fetchMetadata(url string) (string, string, error) {
	cmd := exec.Command("ffmpeg", "-i", url, "-f", "ffmetadata", "-")

	// Capture the output
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	// Run the command
	err := cmd.Run()
	if err != nil {
		return "", "", fmt.Errorf("error running ffmpeg command: %v", err)
	}
	slog.Info("Metadata output:", out.String())
	// Read the metadata output
	var title, artist string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "StreamTitle=") {
			title = strings.TrimPrefix(line, "StreamTitle=")
		} else if strings.HasPrefix(line, "StreamUrl=") {
			artist = strings.TrimPrefix(line, "StreamUrl=")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", "", fmt.Errorf("error reading metadata: %v", err)
	}

	return title, artist, nil
}
