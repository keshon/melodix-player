package discord

import (
	"fmt"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/music/history"
	"github.com/keshon/melodix-discord-player/music/utils"
)

// handleHistoryCommand handles the history command for Discord.
func (d *Discord) handleHistoryCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	d.changeAvatar(s)

	var sortBy string
	var title string

	switch param {
	case "count", "times", "time":
		sortBy, title = "play_count", " — by play count"
	case "duration", "dur":
		sortBy, title = "duration", " — by total duration"
	default:
		sortBy, title = "last_played", " — most recent"
	}

	h := history.NewHistory()
	list, err := h.GetHistory(d.GuildID, sortBy)
	if err != nil {
		slog.Warn("No history table found")
	}

	description := fmt.Sprintf("⏳ **History %v**", title)
	descriptionLength := len(description)

	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	footerLength := len(version.AppFullName)

	for _, elem := range list {
		duration := utils.FormatDuration(elem.History.Duration)
		fieldContent := fmt.Sprintf("```id: %d```    ```count: %d```    ```duration: %v```", elem.History.TrackID, elem.History.PlayCount, duration)
		fieldContentLength := len(fieldContent)
		nameLength := len(elem.Track.Name)
		urlLength := len(elem.Track.URL)

		// Check if adding the field exceeds the embed limit
		if descriptionLength+footerLength+len(embedMsg.Fields)+fieldContentLength+nameLength+urlLength > 4000 {
			break
		}

		// Add the field and update the length counter
		embedMsg.AddField(fieldContent, fmt.Sprintf("[%v](%v)", elem.Track.Name, elem.Track.URL))
		descriptionLength += fieldContentLength + nameLength
	}
	fmt.Println(descriptionLength)
	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
	if err != nil {
		slog.Warnf("Error sending history message: %v", err)
	}
}
