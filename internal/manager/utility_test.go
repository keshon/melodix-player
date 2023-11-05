package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		input         string
		expectedCmd   string
		expectedParam string
	}{
		{"play song", "play", "song"},
		{"pause  ", "pause", " "},
	}

	for _, test := range tests {
		cmd, param, _ := parseCommand(test.input, prefix)
		assert.Equal(t, test.expectedCmd, cmd, "Unexpected command for input: %s", test.input)
		assert.Equal(t, test.expectedParam, param, "Unexpected parameter for input: %s", test.input)
	}
}
