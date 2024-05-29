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

func (d *Discord) handleNowPlayngCommand() {
	if d.Player.GetCurrentStatus() != player.StatusPlaying {
		d.sendMessageEmbed("Player must be playing first")
		return
	}

	var title string
	var err error

	if d.Player.GetCurrentSong().Source == player.SourceStream {
		url := d.Player.GetCurrentSong().Filepath
		title, err = fetchMetadata(url)
		if err != nil {
			d.sendMessageEmbed(err.Error())
			return
		}
	} else {
		title = d.Player.GetCurrentSong().Title
	}

	d.sendMessageEmbed(fmt.Sprintf("```%s```", title))
}

func fetchMetadata(url string) (string, error) {
	cmd := exec.Command("ffmpeg", "-i", url, "-f", "ffmetadata", "-")

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("error running ffmpeg command: %v", err)
	}

	slog.Info("Metadata output:", out.String())

	var title string
	scanner := bufio.NewScanner(&out)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "StreamTitle=") {
			title = strings.TrimPrefix(line, "StreamTitle=")
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading metadata: %v", err)
	}

	return title, nil
}
