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
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// parseInt64 parses the input string into a 64-bit signed integer.
//
// It takes a string as a parameter and returns a 64-bit signed integer and an error.
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// parseFloat64 parses the input string as a 64-bit floating point number.
//
// It takes a string as input and returns a float64 and an error.
func parseFloat64(s string) (float64, error) {
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
func min(a, b int) int {
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
func absDiffInt(x, y int) int {
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
func absDiffUint(x, y uint) uint {
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

	videoInfo.CNR, err = parseInt(queryParams.Get("cnr"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cnr: %v", err)
	}

	videoInfo.Duration, err = parseFloat64(queryParams.Get("dur"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse duration: %v", err)
	}

	videoInfo.EI = queryParams.Get("ei")

	videoInfo.Expire, err = parseInt64(queryParams.Get("expire"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse expire: %v", err)
	}

	videoInfo.FExp, err = parseInt(queryParams.Get("fexp"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse fexp: %v", err)
	}

	videoInfo.FVIP, err = parseInt(queryParams.Get("fvip"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse fvip: %v", err)
	}
	videoInfo.ID = queryParams.Get("id")

	videoInfo.InitCwndBPS, err = parseInt(queryParams.Get("initcwndbps"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse initcwndbps: %v", err)
	}

	videoInfo.IP = queryParams.Get("ip")

	videoInfo.Itag, err = parseInt(queryParams.Get("itag"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse itag: %v", err)
	}

	videoInfo.LMT, err = parseInt64(queryParams.Get("lmt"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse lmt: %v", err)
	}

	videoInfo.LSig = queryParams.Get("lsig")
	videoInfo.LSParams = queryParams.Get("lsparams")
	videoInfo.MH = queryParams.Get("mh")
	videoInfo.MIME = queryParams.Get("mime")
	videoInfo.MM = queryParams.Get("mm")
	videoInfo.MN = queryParams.Get("mn")
	videoInfo.MS = queryParams.Get("ms")

	videoInfo.MT, err = parseInt64(queryParams.Get("mt"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mt: %v", err)
	}

	videoInfo.MV = queryParams.Get("mv")

	videoInfo.MVI, err = parseInt(queryParams.Get("mvi"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse mvi: %v", err)
	}

	videoInfo.PL, err = parseInt(queryParams.Get("pl"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse pl: %v", err)
	}

	videoInfo.RateBypass = queryParams.Get("ratebypass")
	videoInfo.RequireSSL = queryParams.Get("requiressl")
	videoInfo.Sig = queryParams.Get("sig")
	videoInfo.Source = queryParams.Get("source")
	videoInfo.SParams = queryParams.Get("sparams")

	videoInfo.SVPUC, err = parseInt(queryParams.Get("svpuc"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse svpuc: %v", err)
	}

	videoInfo.TXP, err = parseInt(queryParams.Get("txp"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse txp: %v", err)
	}

	videoInfo.VPRV, err = parseInt(queryParams.Get("vprv"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse vprv: %v", err)
	}

	videoInfo.XPC = queryParams.Get("xpc")

	return videoInfo, nil
}