package utils

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/bwmarrin/discordgo"
)

// FormatDuration formats the given duration in seconds into a string in the format "hh:mm:ss".
//
// seconds float64 - the duration in seconds to be formatted
// string - the formatted duration string
func FormatDuration(seconds float64) string {
	h := int(seconds) / 3600
	m := int(seconds) % 3600 / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

// ReadFileToBase64 reads the content of the file at the given filePath and returns it as a base64 encoded string.
//
// Parameter:
// filePath string - the path to the file to be read.
// Return:
// string - the base64 encoded content of the file.
// error - an error if any operation fails.
func ReadFileToBase64(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("error opening the file: %v", err)
	}
	defer file.Close()

	stat, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("error getting file information: %v", err)
	}

	fileContent := make([]byte, stat.Size())
	_, err = file.Read(fileContent)
	if err != nil {
		return "", fmt.Errorf("error reading the file: %v", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileContent)
	return fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(fileContent), base64Content), nil
}

// SanitizeString sanitizes the input string by removing non-printable unicode characters.
//
// It takes a string input and returns a sanitized string.
func SanitizeString(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, input)
}

// InferProtocolByPort is a function that infers the protocol based on the hostname and port.
//
// It takes a hostname of type string and a port of type int as parameters, and returns a string.
func InferProtocolByPort(hostname string, port int) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return "http://"
	}
	defer conn.Close()
	return "https://"
}

// parseInt parses the input string to an integer.
//
// It takes a string as input and returns an integer and an error.
func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseInt64 parses the input string into a 64-bit signed integer.
//
// It takes a string as a parameter and returns a 64-bit signed integer and an error.
func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// parseFloat64 parses the input string as a 64-bit floating point number.
//
// It takes a string as input and returns a float64 and an error.
func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// GetRandomImagePathFromPath returns a random image path from the given folder path.
//
// It takes a folderPath string as a parameter and returns a string and an error.
func GetRandomImagePathFromPath(folderPath string) (string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	var validFiles []string
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			validFiles = append(validFiles, file.Name())
		}
	}

	if len(validFiles) == 0 {
		return "", fmt.Errorf("no valid images found")
	}

	randomIndex := rand.Intn(len(validFiles))
	randomImage := validFiles[randomIndex]
	imagePath := filepath.Join(folderPath, randomImage)

	return imagePath, nil
}

// GetWeightedRandomImagePath returns a random image path with reduced chances for recently shown images.
func GetWeightedRandomImagePath(folderPath string) (string, error) {
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	var images []string
	var weights []int

	for _, file := range files {
		if file.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(file.Name()))
		if ext == ".jpg" || ext == ".jpeg" || ext == ".png" {
			images = append(images, file.Name())
			weights = append(weights, 1) // Initial weight for each image is 1
		}
	}

	if len(images) == 0 {
		return "", fmt.Errorf("no valid images found")
	}

	totalWeights := len(images) // Initial total weights equal to the number of images

	randomWeight := rand.Intn(totalWeights)

	index := -1
	for i, weight := range weights {
		if randomWeight < weight {
			index = i
			break
		}
		randomWeight -= weight
	}

	if index == -1 {
		return "", fmt.Errorf("error selecting random image")
	}

	// Decrease the weight of the recently selected image
	weights[index] = weights[index] / 2

	// Increase the weight of all other images
	for i := range weights {
		if i != index {
			weights[i] = weights[i] * 2
		}
	}

	imagePath := filepath.Join(folderPath, images[index])
	return imagePath, nil
}

// TrimString trims the input string to the specified limit.
//
// It takes a string input and an integer limit as parameters and returns a string.
func TrimString(input string, limit int) string {
	return input[:min(len(input), limit)]
}

// min returns the minimum of two integers.
//
// Parameters:
//
//	a int - the first integer
//	b int - the second integer
//
// Return type:
//
//	int - the minimum of the two integers
func Min(a, b int) int {
	return a&((a-b)>>31) | b&(^((a - b) >> 31))
}

// AbsInt returns the absolute value of an integer.
//
// x: the input integer
// int: the absolute value of the input integer
func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// absDiffInt calculates the absolute difference between two integers.
//
// x, y are the integers to find the absolute difference between.
// int is the return type, representing the absolute difference.
func AbsDiffInt(x, y int) int {
	diff := x - y
	if diff < 0 {
		return -diff
	}
	return diff
}

// absDiffUint calculates the absolute difference between two unsigned integers.
//
// x, y uint
// uint
func AbsDiffUint(x, y uint) uint {
	if x < y {
		return y - x
	}
	return x - y
}

// findUserVoiceState finds the voice state of the user in the given list of voice states.
//
// userID: The ID of the user to find the voice state for.
// voiceStates: The list of voice states to search through.
// *discordgo.VoiceState, bool: The found voice state and a boolean indicating if it was found.
func findUserVoiceState(userID string, voiceStates []*discordgo.VoiceState) (foundState *discordgo.VoiceState, found bool) {
	for _, state := range voiceStates {
		if state != nil && state.UserID == userID {
			foundState = state
			found = true
			return
		}
	}
	return
}

func IsYouTubeURL(url string) bool {
	pattern := regexp.MustCompile(`^(https?://)?(www\.)?(youtube\.com|youtu\.be)/.*$`)
	return pattern.MatchString(strings.ToLower(url))
}

func IsAudioFile(fileName string) bool {
	fileName = strings.ToLower(fileName)

	audioExtensions := []string{".mp3", ".wav", ".ogg", ".flac", ".aac", ".m4a"}

	for _, ext := range audioExtensions {
		if strings.HasSuffix(fileName, ext) {
			return true
		}
	}

	return false
}

func IsValidHttpUrl(u string) bool {
	_, err := url.Parse(u)
	return err == nil
}
