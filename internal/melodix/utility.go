package melodix

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// parseCommand parses the command and parameter from the Discord input based on the provided pattern.
func parseCommand(content, pattern string) (string, string, error) {
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
func getCanonicalCommand(alias string, commandAliases [][]string) string {
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
func parseSongsAndTypeInParameter(param string) (string, []string) {
	// Trim spaces at the beginning and end
	param = strings.TrimSpace(param)

	if len(param) == 0 {
		return "", []string{}
	}

	// Check if the parameter is a URL
	u, err := url.Parse(param)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && isYouTubeURL(u.Host) {
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
func isYouTubeURL(host string) bool {
	return host == "www.youtube.com" || host == "youtube.com" || host == "youtu.be"
}

// parseVideoParamsFromYoutubeURL parses video parameters from a YouTube URL.
func parseVideoParamsFromYouTubeURL(urlString string) (duration float64, contentLength int, expiryTimestamp int64, err error) {
	duration = -1
	contentLength = -1
	expiryTimestamp = -1

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return duration, contentLength, expiryTimestamp, fmt.Errorf("failed to parse URL: %v", err)
	}

	queryParams := parsedURL.Query()

	durParam, err := strconv.ParseFloat(queryParams.Get("dur"), 64)
	if err != nil {
		duration = -1
	}
	duration = durParam

	if clenParam := queryParams.Get("clen"); clenParam != "" {
		contentLength, err = strconv.Atoi(clenParam)
		if err != nil {
			return duration, contentLength, expiryTimestamp, fmt.Errorf("failed to parse content length: %v", err)
		}
	}

	if expireParam := queryParams.Get("expire"); expireParam != "" {
		expiryTimestamp, err = strconv.ParseInt(expireParam, 10, 64)
		if err != nil {
			return duration, contentLength, expiryTimestamp, fmt.Errorf("failed to parse expiry timestamp: %v", err)
		}
	}

	return duration, contentLength, expiryTimestamp, nil
}

// String returns the string representation of the CurrentStatus.
func (status Status) String() string {
	statuses := map[Status]string{
		StatusResting: "Resting",
		StatusPlaying: "Playing",
		StatusPaused:  "Paused",
		StatusError:   "Error",
	}

	return statuses[status]
}

// formatDuration formats the given seconds into HH:MM:SS format.
func formatDuration(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	totalSeconds %= 3600
	minutes := totalSeconds / 60
	seconds = math.Mod(float64(totalSeconds), 60)
	return fmt.Sprintf("%02d:%02d:%02.0f", hours, minutes, seconds)
}

// getRandomAvatarPath returns path to randomly selected file in specified folder
func getRandomImagePathFromPath(folderPath string) (string, error) {

	var validFiles []string
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	// Filter only files with certain extensions (you can modify this if needed)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".png" {
			validFiles = append(validFiles, file.Name())
		}
	}

	if len(validFiles) == 0 {
		return "", fmt.Errorf("no valid images found")
	}

	// Get a random index
	randomIndex := rand.Intn(len(validFiles))
	randomImage := validFiles[randomIndex]
	imagePath := filepath.Join(folderPath, randomImage)

	return imagePath, nil
}

func readFileToBase64(imgPath string) (string, error) {
	img, err := os.ReadFile(imgPath)
	if err != nil {
		return "", fmt.Errorf("error reading the response: %v", err)
	}

	base64Img := base64.StdEncoding.EncodeToString(img)
	return fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64Img), nil
}

func getRandomWaitPhrase() string {
	phrases := []string{
		"Chillax, I'm on it...",
		"Easy there, turbo...",
		"Ever heard of fashionably late?",
		"Hold your horses, we got this...",
		"Patience, my young padawan...",
		"I move at my own pace, deal with it...",
		"Slow and steady wins the race, right?",
		"Taking my time, just like a fine wine...",
		"Can't rush perfection, my friend...",
		"Grab a snack, this might take a minute...",
		"Tick-tock, but in my own clock...",
		"Did someone order a chilled response?",
		"Sit back, relax, and enjoy the show...",
		"Don't rush me, I'm on island time...",
		"Mastering the art of fashionably late...",
		"Patience, grasshopper...",
		"Hang in there, superstar...",
		"Hold my server, I got this...",
		"Data's doing the cha-cha...",
		"Server's got moves, wait...",
		"Code's flexing its muscles...",
		"Binary bits breakdancing...",
		"Servers tap dancing for you...",
		"Coding wizardry in progress...",
		"Request on a magic carpet...",
		"Cyber monkeys typing furiously...",
		"Your wish is my command...almost...",
		"Quantum computing, almost there...",
		"Data sprinting to your screen...",
		"Virtual acrobatics in motion...",
		"Code juggling like a boss...",
		"Bytes breakdancing in the server...",
		"Request breakdancing through firewalls...",
		"Code tap dancing its way...",
		"Server's telling knock-knock jokes...",
		"Request on a virtual rollercoaster...",
		"Algorithms breakdancing for you...",
		"Ninja moves on your request...",
		"Coffee break while we work...",
		"Request moonwalking to completion...",
		"Wild times in the server room...",
		"Sit back, enjoy the show...",
		"Sloth could be faster, but we're on it...",
		"Grab popcorn, it's interesting...",
		"Your request is the VIP...",
		"Put on a seatbelt, bumpy ride...",
		"Request on a data rollercoaster...",
		"Cha-cha with our servers...",
		"Counting to infinity... almost done...",
		"Brace yourself, request is dropping...",
		"Working harder than a cat...",
		"Fairy dust, request complete...",
		"Hold on tight, breakdancing to you...",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}
