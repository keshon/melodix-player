package melodix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	prefix := "!"

	tests := []struct {
		input         string
		expectedCmd   string
		expectedParam string
	}{
		{"!play song", "play", "song"},
		{"!pause  ", "pause", ""},
	}

	for _, test := range tests {
		cmd, param, _ := parseCommand(test.input, prefix)
		assert.Equal(t, test.expectedCmd, cmd, "Unexpected command for input: %s", test.input)
		assert.Equal(t, test.expectedParam, param, "Unexpected parameter for input: %s", test.input)
	}
}

func TestStatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusResting, "Resting"},
		{StatusPlaying, "Playing"},
		{StatusPaused, "Paused"},
		{StatusError, "Error"},
	}

	for _, test := range tests {
		result := test.status.String()
		assert.Equal(t, test.expected, result, "Unexpected status string for status: %d", test.status)
	}
}
