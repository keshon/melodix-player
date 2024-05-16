package utils

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

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

func TrimString(input string, limit int) string {
	return input[:min(len(input), limit)]
}

func min(a, b int) int {
	return a&((a-b)>>31) | b&(^((a - b) >> 31))
}

func AbsInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
