package utils

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

// formatDuration formats the given seconds into HH:MM:SS format.
func FormatDuration(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	totalSeconds %= 3600
	minutes := totalSeconds / 60
	seconds = math.Mod(float64(totalSeconds), 60)
	return fmt.Sprintf("%02d:%02d:%02.0f", hours, minutes, seconds)
}

func ReadFileToBase64(imgPath string) (string, error) {
	img, err := os.ReadFile(imgPath)
	if err != nil {
		return "", fmt.Errorf("error reading the response: %v", err)
	}

	base64Img := base64.StdEncoding.EncodeToString(img)
	return fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64Img), nil
}

func SanitizeString(input string) string {
	// Define a regular expression to match unwanted characters
	re := regexp.MustCompile("[[:^print:]]")

	// Replace unwanted characters with an empty string
	sanitized := re.ReplaceAllString(input, "")

	return sanitized
}

// inferProtocolByPort attempts to infer the protocol based on the availability of a specific port.
func InferProtocolByPort(hostname string, port int) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		// Assuming it's not available, use HTTP
		return "http://"
	}
	defer conn.Close()

	// The port is available, use HTTPS
	return "https://"
}

// parseVideoParamsFromYoutubeURL parses video parameters from a YouTube URL.
func ParseVideoParamsFromYouTubeURL(urlString string) (duration float64, contentLength int, expiryTimestamp int64, err error) {
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

// getRandomAvatarPath returns path to randomly selected file in specified folder
func GetRandomImagePathFromPath(folderPath string) (string, error) {

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
