package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewConfig(t *testing.T) {
	// Create a temporary environment variable for testing
	os.Setenv("DISCORD_COMMAND_PREFIX", "!")
	os.Setenv("DISCORD_BOT_TOKEN", "example-token")
	os.Setenv("REST_ENABLED", "true")
	os.Setenv("REST_GIN_RELEASE", "false")
	os.Setenv("REST_HOSTNAME", "example-hostname")

	// Clean up the environment variables when the test is done
	defer func() {
		os.Clearenv()
	}()

	// Test the NewConfig function
	cfg, err := NewConfig()
	assert.NoError(t, err, "NewConfig should not return an error")

	// Validate the Config object
	assert.Equal(t, "!", cfg.DiscordCommandPrefix, "DiscordCommandPrefix should match")
	assert.Equal(t, "example-token", cfg.DiscordBotToken, "DiscordBotToken should match")
	assert.True(t, cfg.RestEnabled, "RestEnabled should be true")
	assert.False(t, cfg.RestGinRelease, "RestGinRelease should be false")
	assert.Equal(t, "example-hostname", cfg.RestHostname, "RestHostname should match")
}

func TestString(t *testing.T) {
	// Create a Config object for testing
	cfg := &Config{
		DiscordCommandPrefix: "!",
		DiscordBotToken:      "example-token",
		RestEnabled:          true,
		RestGinRelease:       false,
		RestHostname:         "example-hostname",
	}

	// Test the String method
	result := cfg.String()

	// Define the expected JSON representation
	expectedJSON := `{
        "DiscordCommandPrefix": "!",
        "DiscordBotToken": "example-token",
        "RestEnabled": true,
        "RestGinRelease": false,
        "RestHostname": "example-hostname"
    }`

	// Validate the JSON representation
	assert.JSONEq(t, expectedJSON, result, "String representation should match expected JSON")
}

func TestValidateMandatoryConfig(t *testing.T) {
	// Test case: All mandatory environment variables are set
	os.Setenv("DISCORD_COMMAND_PREFIX", "!")
	os.Setenv("DISCORD_BOT_TOKEN", "example-token")
	os.Setenv("REST_ENABLED", "true")

	err := validateMandatoryConfig()
	assert.NoError(t, err, "No error should be returned when all mandatory variables are set")

	// Test case: One mandatory environment variable is missing
	os.Unsetenv("DISCORD_COMMAND_PREFIX")

	err = validateMandatoryConfig()
	assert.Error(t, err, "An error should be returned when a mandatory variable is missing")
	assert.Contains(t, err.Error(), "DISCORD_COMMAND_PREFIX", "Error message should mention the missing variable")
}
