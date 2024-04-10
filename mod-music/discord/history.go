package discord

import (
	"fmt"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/mod-music/history"
	"github.com/keshon/melodix-player/mod-music/utils"
)

func (d *Discord) handleHistoryCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	d.changeAvatar(s)

	sortBy, title := "last_played", " — most recent"
	switch param {
	case "count", "times", "time":
		sortBy, title = "play_count", " — by play count"
	case "duration", "dur":
		sortBy, title = "duration", " — by total duration"
	}

	historyManager := history.NewHistory()
	historyList, err := historyManager.GetHistory(d.GuildID, sortBy)
	if err != nil {
		slog.Error("Error retrieving history", err)
		return
	}

	description := fmt.Sprintf("⏳ History %v", title)
	if len(description) > 4096 {
		description = utils.TrimString(description, 4096)
	}

	embedMsg := embed.NewEmbed().
		SetDescription(description).
		SetColor(0x9f00d4)

	maxLimit := 6000 - len(description)

	for i, elem := range historyList {
		if i > 24 {
			break
		}

		duration := utils.FormatDuration(elem.History.Duration)
		fieldContent := fmt.Sprintf("```id %d```\t```x%d```\t```%v```\t```%v```", elem.History.TrackID, elem.History.PlayCount, duration, strings.ToLower(elem.Track.Source))

		if remainingSpace := maxLimit - len(embedMsg.Fields) - len(fieldContent) - len(elem.Track.Title) - len(elem.Track.URL); remainingSpace < 0 {
			break
		}

		if elem.Track.URL != "" {
			embedMsg.AddField(fieldContent, fmt.Sprintf("[%v](%v)\n\n", elem.Track.Title, elem.Track.URL))
		} else {
			embedMsg.AddField(fieldContent, fmt.Sprintf("%v\n\n", elem.Track.Title))
		}
	}

	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
	if err != nil {
		slog.Error("Error sending history message", err)
	}
}
