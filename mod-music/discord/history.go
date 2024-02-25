package discord

import (
	"fmt"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
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

	description := fmt.Sprintf("⏳ History %v", title)
	if len(description) > 4096 {
		description = utils.TrimString(description, 4096)
	}
	descriptionLength := len(description)

	embedMsg := embed.NewEmbed().
		SetDescription(description).
		SetColor(0x9f00d4)

	maxLimit := 6000 - descriptionLength

	for i, elem := range list {
		if i > 24 {
			break
		}

		duration := utils.FormatDuration(elem.History.Duration)
		fieldContent := fmt.Sprintf("```id: %d```    ```count: %d```    ```duration: %v```", elem.History.TrackID, elem.History.PlayCount, duration)
		fieldContentLength := len(fieldContent)

		nameLength := len(elem.Track.Name)
		urlLength := len(elem.Track.URL)

		if maxLimit-len(embedMsg.Fields)-fieldContentLength-nameLength-urlLength < 0 {
			break
		}

		embedMsg.AddField(fieldContent, fmt.Sprintf("[%v](%v)\n▬▬▬▬▬▬▬▬▬▬\n", elem.Track.Name, elem.Track.URL))
	}

	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
	if err != nil {
		slog.Warnf("Error sending history message: %v", err)
	}
}
