package config

import (
	"encoding/json"
	"errors"
	"log/slog"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/keshon/melodix-player/mods/music/third_party/dca"
)

type Config struct {
	DiscordCommandPrefix       string
	DiscordBotToken            string
	DiscordAdminUserID         string
	DiscordAdminChannelID      string
	RestEnabled                bool
	RestGinRelease             bool
	RestHostname               string
	DcaFrameDuration           int
	DcaBitrate                 int
	DcaPacketLoss              int
	DcaRawOutput               bool
	DcaApplication             dca.AudioApplication
	DcaCompressionLevel        int
	DcaBufferedFrames          int
	DcaVBR                     bool
	DcaReconnectAtEOF          int // boolean value passed to Ffmpeg is treated as int (1 - true, 0 - false)
	DcaReconnectStreamed       int // boolean value passed to Ffmpeg is treated as int (1 - true, 0 - false)
	DcaReconnectOnNetworkError int // boolean value passed to Ffmpeg is treated as int (1 - true, 0 - false)
	DcaReconnectOnHttpError    string
	DcaReconnectDelayMax       int
	DcaFfmpegBinaryPath        string
	DcaEncodingLineLog         bool
	DcaUserAgent               string
}

func NewConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	if err := validateMandatoryConfig(); err != nil {
		return nil, err
	}

	config := &Config{
		DiscordCommandPrefix:       os.Getenv("DISCORD_COMMAND_PREFIX"),
		DiscordBotToken:            os.Getenv("DISCORD_BOT_TOKEN"),
		DiscordAdminUserID:         os.Getenv("DISCORD_ADMIN_USER_ID"),
		DiscordAdminChannelID:      os.Getenv("DISCORD_ADMIN_CHANNEL_ID"),
		RestEnabled:                getenvAsBool("REST_ENABLED"),
		RestGinRelease:             getenvAsBool("REST_GIN_RELEASE"),
		RestHostname:               os.Getenv("REST_HOSTNAME"),
		DcaFrameDuration:           getenvAsInt("DCA_FRAME_DURATION"),
		DcaBitrate:                 getenvAsInt("DCA_BITRATE"),
		DcaPacketLoss:              getenvAsInt("DCA_PACKET_LOSS"),
		DcaRawOutput:               getenvAsBool("DCA_RAW_OUTPUT"),
		DcaApplication:             dca.AudioApplication(os.Getenv("DCA_APPLICATION")),
		DcaCompressionLevel:        getenvAsInt("DCA_COMPRESSION_LEVEL"),
		DcaBufferedFrames:          getenvAsInt("DCA_BUFFERED_FRAMES"),
		DcaVBR:                     getenvAsBool("DCA_VBR"),
		DcaReconnectAtEOF:          getenvBoolAsInt("DCA_RECONNECT_AT_EOF"),
		DcaReconnectStreamed:       getenvBoolAsInt("DCA_RECONNECT_STREAMED"),
		DcaReconnectOnNetworkError: getenvBoolAsInt("DCA_RECONNECT_ON_NETWORK_ERROR"),
		DcaReconnectOnHttpError:    os.Getenv("DCA_RECONNECT_ON_HTTTP_ERROR"),
		DcaReconnectDelayMax:       getenvAsInt("DCA_RECONNECT_MAX"),
		DcaFfmpegBinaryPath:        os.Getenv("DCA_FFMPEG_BINARY_PATH"),
		DcaEncodingLineLog:         getenvAsBool("DCA_ENCODING_LINE_LOG"),
		DcaUserAgent:               os.Getenv("DCA_USER_AGENT"),
	}

	return config, nil
}

// Returns the JSON representation of the Config struct.
func (c *Config) String() string {
	configMap := map[string]interface{}{
		"DiscordCommandPrefix":       c.DiscordCommandPrefix,
		"DiscordBotToken":            c.DiscordBotToken,
		"DiscordAdminUserID":         c.DiscordAdminUserID,
		"DiscordAdminChannelID":      c.DiscordAdminChannelID,
		"RestEnabled":                c.RestEnabled,
		"RestGinRelease":             c.RestGinRelease,
		"RestHostname":               c.RestHostname,
		"DcaFrameDuration":           c.DcaFrameDuration,
		"DcaBitrate":                 c.DcaBitrate,
		"DcaPacketLoss":              c.DcaPacketLoss,
		"DcaRawOutput":               c.DcaRawOutput,
		"DcaApplication":             c.DcaApplication,
		"DcaCompressionLevel":        c.DcaCompressionLevel,
		"DcaBufferedFrames":          c.DcaBufferedFrames,
		"DcaVBR":                     c.DcaVBR,
		"DcaReconnectAtEOF":          c.DcaReconnectAtEOF,
		"DcaReconnectStreamed":       c.DcaReconnectStreamed,
		"DcaReconnectOnNetworkError": c.DcaReconnectOnNetworkError,
		"DcaReconnectOnHttpError":    c.DcaReconnectOnHttpError,
		"DcaReconnectDelayMax":       c.DcaReconnectDelayMax,
		"DcaFfmpegBinaryPath":        c.DcaFfmpegBinaryPath,
		"DcaEncodingLineLog":         c.DcaEncodingLineLog,
		"DcaUserAgent":               c.DcaUserAgent,
	}

	jsonString, err := json.MarshalIndent(configMap, "", "    ")
	if err != nil {
		return ""
	}

	return string(jsonString)
}

// validateMandatoryConfig checks for the presence of mandatory configuration keys
// in the environment variables and returns an error if any are missing.
func validateMandatoryConfig() error {
	mandatoryKeys := []string{
		"DISCORD_COMMAND_PREFIX", "DISCORD_BOT_TOKEN", "REST_ENABLED", "DCA_FRAME_DURATION", "DCA_BITRATE", "DCA_PACKET_LOSS",
		"DCA_RAW_OUTPUT", "DCA_APPLICATION", "DCA_COMPRESSION_LEVEL", "DCA_BUFFERED_FRAMES",
		"DCA_VBR", "DCA_RECONNECT_AT_EOF", "DCA_RECONNECT_STREAMED", "DCA_RECONNECT_ON_NETWORK_ERROR",
		"DCA_RECONNECT_ON_HTTTP_ERROR", "DCA_RECONNECT_MAX",
		"DCA_ENCODING_LINE_LOG", "DCA_USER_AGENT",
	}

	for _, key := range mandatoryKeys {
		if os.Getenv(key) == "" {
			return errors.New("Missing mandatory configuration: " + key)
		}
	}

	return nil
}

func getenvAsBool(key string) bool {
	val := os.Getenv(key)

	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		return false
	}

	return boolValue
}

func getenvAsInt(key string) int {
	val := os.Getenv(key)

	intValue, err := strconv.Atoi(val)
	if err != nil {
		slog.Error("Error parsing integer value from env variable")
		return 0
	}

	return intValue
}

func getenvBoolAsInt(key string) int {
	val := os.Getenv(key)

	boolValue, err := strconv.ParseBool(val)
	if err != nil {
		slog.Error("Error parsing bool value from env variable")
		return 0
	}

	if boolValue {
		return 1
	}

	return 0
}
