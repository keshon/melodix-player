package discord

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/music/utils"
)

// parseCommand parses the command and parameter from the Discord input based on the provided pattern.
func ParseCommand(content, pattern string) (string, string, error) {
	if !strings.HasPrefix(content, pattern) {
		return "", "", fmt.Errorf("pattern not found")
	}

	content = content[len(pattern):] // Strip the pattern

	words := strings.Fields(content) // Split by whitespace, handling multiple spaces
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := strings.ToLower(words[0])
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}

// getCanonicalCommand gets the canonical command from aliases using the given alias.
func GetCanonicalCommand(alias string, commandAliases [][]string) string {
	for _, aliases := range commandAliases {
		for _, a := range aliases {
			if a == alias {
				return aliases[0]
			}
		}
	}
	return ""
}

// parseSongsAndTypeInParameter parses the type and parameters from the input parameter string.
func ParseSongsAndTypeInParameter(param string) (string, []string) {
	// Trim spaces at the beginning and end
	param = strings.TrimSpace(param)

	if len(param) == 0 {
		return "", []string{}
	}

	// Check if the parameter is a URL
	u, err := url.Parse(param)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && IsYouTubeURL(u.Host) {
		// If it's a URL, split by ",", " ", new line, or carriage return
		paramSlice := strings.FieldsFunc(param, func(r rune) bool {
			return r == ',' || r == '\n' || r == '\r' || r == ' ' || r == '\t'
		})
		return "url", paramSlice
	}

	// Check if the parameter is an ID
	params := strings.Fields(param)
	allValidIDs := true
	for _, param := range params {
		_, err := strconv.Atoi(param)
		if err != nil {
			allValidIDs = false
			break
		}
	}
	if allValidIDs {
		return "id", params
	}

	// Treat it as a single title if it's not a URL or ID
	return "title", []string{param}
}

// isYouTubeURL checks if the host is a YouTube URL.
func IsYouTubeURL(host string) bool {
	return host == "www.youtube.com" || host == "youtube.com" || host == "youtu.be"
}

// createThumbnail returns encoded randomly picked avatar image
func createThumbnail() (string, error) {
	imgPath, err := utils.GetRandomImagePathFromPath("./assets/avatars")
	if err != nil {
		slog.Errorf("Error getting avatar path: %v", err)
		return "", err
	}

	avatar, err := utils.ReadFileToBase64(imgPath)
	if err != nil {
		fmt.Printf("Error preparing avatar: %v\n", err)
		return "", err
	}

	return avatar, nil
}

// changeAvatar changes bot avatar with randomly picked avatar image within allowed rate limit
func (d *Discord) changeAvatar(s *discordgo.Session) {
	// Check if the rate limit duration has passed since the last execution
	if time.Since(d.lastChangeAvatarTime) < d.rateLimitDuration {
		slog.Info("Rate-limited. Skipping changeAvatar.")
		return
	}

	imgPath, err := utils.GetRandomImagePathFromPath("./assets/avatars")
	if err != nil {
		slog.Errorf("Error getting avatar path: %v", err)
		return
	}

	avatar, err := utils.ReadFileToBase64(imgPath)
	if err != nil {
		fmt.Printf("Error preparing avatar: %v\n", err)
		return
	}

	_, err = s.UserUpdate("", avatar)
	if err != nil {
		slog.Errorf("Error setting the avatar: %v", err)
		return
	}

	// Update the last execution time
	d.lastChangeAvatarTime = time.Now()
}

func findUserVoiceState(userID string, voiceStates []*discordgo.VoiceState) (*discordgo.VoiceState, bool) {
	for _, vs := range voiceStates {
		if vs.UserID == userID {
			return vs, true
		}
	}
	return nil, false
}
