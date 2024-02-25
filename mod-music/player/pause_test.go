package player

import (
	"testing"

	"github.com/bwmarrin/discordgo"
)

func TestPause(t *testing.T) {
	p := NewPlayer("123456")

	// Test pause when there is no voice connection
	p.Pause()

	if p.GetCurrentStatus() != StatusResting {
		t.Error("Expected current status to remain as StatusResting")
	}

	// Test pause when there is a voice connection but no streaming session
	p.SetVoiceConnection(&discordgo.VoiceConnection{})
	p.Pause()

	if p.GetCurrentStatus() != StatusResting {
		t.Error("Expected current status to remain as StatusResting")
	}
}
