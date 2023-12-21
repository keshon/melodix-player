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
)

// FormatDuration formats the given seconds into HH:MM:SS format.
// Example: formattedTime := FormatDuration(3661.5) // Returns "01:01:02"
func FormatDuration(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	totalSeconds %= 3600
	minutes := totalSeconds / 60
	seconds = math.Mod(float64(totalSeconds), 60)
	return fmt.Sprintf("%02d:%02d:%02.0f", hours, minutes, seconds)
}

// ReadFileToBase64 reads a file and returns its base64 representation with data URI.
// Example: base64Data, err := ReadFileToBase64("/path/to/image.jpg")
func ReadFileToBase64(filePath string) (string, error) {
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("error reading the file: %v", err)
	}

	base64Content := base64.StdEncoding.EncodeToString(fileContent)
	return fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(fileContent), base64Content), nil
}

// SanitizeString removes unwanted characters from the input string.
// Example: sanitizedStr := SanitizeString("Hello#World!")
func SanitizeString(input string) string {
	unwantedCharRegex := regexp.MustCompile("[[:^print:]]")
	return unwantedCharRegex.ReplaceAllString(input, "")
}

// InferProtocolByPort attempts to infer the protocol based on the availability of a specific port.
// Example: protocol := InferProtocolByPort("example.com", 443)
func InferProtocolByPort(hostname string, port int) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		// Assuming it's not available, default to HTTP
		return "http://"
	}
	defer conn.Close()

	// The port is available, use HTTPS
	return "https://"
}

// ParseQueryParamsFromURL parses query parameters from a URL.
// Example: params, err := ParseQueryParamsFromURL("https://www.example.com/path?param1=value1&param2=value2")
func ParseQueryParamsFromURL(urlString string) (map[string]string, error) {
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	queryParams := parsedURL.Query()

	params := make(map[string]string)
	for key, values := range queryParams {
		// For simplicity, consider only the first value if there are multiple values for a key
		params[key] = values[0]
	}

	return params, nil
}

// GetRandomImagePathFromPath returns the path to a randomly selected file in the specified folder.
// Example: imagePath, err := GetRandomImagePathFromPath("/path/to/images")
func GetRandomImagePathFromPath(folderPath string) (string, error) {
	var validFiles []string
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	// Filter only files with certain extensions (you can modify this if needed)
	for _, file := range files {
		if ext := filepath.Ext(file.Name()); ext == ".jpg" || ext == ".png" {
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

// TrimString trims the string's ending beyond the specified character limit.
// Example: trimmedText := TrimString("This is a long text.", 10) // Returns "This is a"
func TrimString(input string, limit int) string {
	if len(input) <= limit {
		return input
	}

	return input[:limit]
}
