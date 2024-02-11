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

// VideoInfo struct represents the map response
type VideoInfo struct {
	C           string  `json:"c"`
	CNR         int     `json:"cnr"`
	Duration    float64 `json:"dur"`
	EI          string  `json:"ei"`
	Expire      int64   `json:"expire"`
	FExp        int     `json:"fexp"`
	FVIP        int     `json:"fvip"`
	ID          string  `json:"id"`
	InitCwndBPS int     `json:"initcwndbps"`
	IP          string  `json:"ip"`
	Itag        int     `json:"itag"`
	LMT         int64   `json:"lmt"`
	LSig        string  `json:"lsig"`
	LSParams    string  `json:"lsparams"`
	MH          string  `json:"mh"`
	MIME        string  `json:"mime"`
	MM          string  `json:"mm"`
	MN          string  `json:"mn"`
	MS          string  `json:"ms"`
	MT          int64   `json:"mt"`
	MV          string  `json:"mv"`
	MVI         int     `json:"mvi"`
	PL          int     `json:"pl"`
	RateBypass  string  `json:"ratebypass"`
	RequireSSL  string  `json:"requiressl"`
	Sig         string  `json:"sig"`
	Source      string  `json:"source"`
	SParams     string  `json:"sparams"`
	SVPUC       int     `json:"svpuc"`
	TXP         int     `json:"txp"`
	VPRV        int     `json:"vprv"`
	XPC         string  `json:"xpc"`
}

// ParseQueryParamsFromURL parses query parameters from a URL.
// Example: params, err := ParseQueryParamsFromURL("https://www.example.com/path?param1=value1&param2=value2")
func ParseQueryParamsFromURL(urlString string) (*VideoInfo, error) {

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %v", err)
	}

	queryParams := parsedURL.Query()

	videoInfo := &VideoInfo{}
	videoInfo.C = queryParams.Get("c")
	videoInfo.CNR = parseInt(queryParams.Get("cnr"))
	videoInfo.Duration = parseFloat(queryParams.Get("dur"))
	videoInfo.EI = queryParams.Get("ei")
	videoInfo.Expire = parseInt64(queryParams.Get("expire"))
	videoInfo.FExp = parseInt(queryParams.Get("fexp"))
	videoInfo.FVIP = parseInt(queryParams.Get("fvip"))
	videoInfo.ID = queryParams.Get("id")
	videoInfo.InitCwndBPS = parseInt(queryParams.Get("initcwndbps"))
	videoInfo.IP = queryParams.Get("ip")
	videoInfo.Itag = parseInt(queryParams.Get("itag"))
	videoInfo.LMT = parseInt64(queryParams.Get("lmt"))
	videoInfo.LSig = queryParams.Get("lsig")
	videoInfo.LSParams = queryParams.Get("lsparams")
	videoInfo.MH = queryParams.Get("mh")
	videoInfo.MIME = queryParams.Get("mime")
	videoInfo.MM = queryParams.Get("mm")
	videoInfo.MN = queryParams.Get("mn")
	videoInfo.MS = queryParams.Get("ms")
	videoInfo.MT = parseInt64(queryParams.Get("mt"))
	videoInfo.MV = queryParams.Get("mv")
	videoInfo.MVI = parseInt(queryParams.Get("mvi"))
	videoInfo.PL = parseInt(queryParams.Get("pl"))
	videoInfo.RateBypass = queryParams.Get("ratebypass")
	videoInfo.RequireSSL = queryParams.Get("requiressl")
	videoInfo.Sig = queryParams.Get("sig")
	videoInfo.Source = queryParams.Get("source")
	videoInfo.SParams = queryParams.Get("sparams")
	videoInfo.SVPUC = parseInt(queryParams.Get("svpuc"))
	videoInfo.TXP = parseInt(queryParams.Get("txp"))
	videoInfo.VPRV = parseInt(queryParams.Get("vprv"))
	videoInfo.XPC = queryParams.Get("xpc")

	return videoInfo, nil
}

func parseInt(s string) int {
	val, _ := strconv.Atoi(s)
	return val
}

func parseInt64(s string) int64 {
	val, _ := strconv.ParseInt(s, 10, 64)
	return val
}

func parseFloat(s string) float64 {
	val, _ := strconv.ParseFloat(s, 64)
	return val
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
