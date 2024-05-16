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
)

func FormatDurationHHMMSS(seconds float64) string {
	h := int(seconds) / 3600
	m := int(seconds) % 3600 / 60
	s := int(seconds) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

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

func SanitizeString(input string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return -1
	}, input)
}

func InferProtocolByPort(hostname string, port int) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		return "http://"
	}
	defer conn.Close()
	return "https://"
}

func ParseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func ParseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func ParseFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

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

func TrimString(input string, limit int) string {
	return input[:min(len(input), limit)]
}

func Min(a, b int) int {
	return a&((a-b)>>31) | b&(^((a - b) >> 31))
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func AbsDiffInt(x, y int) int {
	diff := x - y
	if diff < 0 {
		return -diff
	}
	return diff
}

func AbsDiffUint(x, y uint) uint {
	if x < y {
		return y - x
	}
	return x - y
}

func IsYouTubeURL(url string) bool {
	pattern := regexp.MustCompile(`^(https?://)?(www\.)?(.*\.)?(youtube\.com|youtu\.be)/.*$`)
	return pattern.MatchString(strings.ToLower(url))
}

func IsValidHttpURL(u string) bool {
	_, err := url.Parse(u)
	return err == nil
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
