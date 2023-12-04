package discord

import (
	"fmt"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/music/player"
)

// handleShowQueueCommand handles the show queue command for Discord.
func (d *Discord) handleShowQueueCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	// Get the current song and playlist
	currentSong := d.Player.GetCurrentSong()
	playlist := d.Player.GetSongQueue()

	// Check if there's a current song or the playlist is not empty
	if currentSong == nil && (len(playlist) == 0) {
		embedMsg.SetDescription("The queue is empty or no current song is playing.")
		s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
		return
	}

	playlistStr := "📑 **The queue**\n\n"

	var newPlaylist []*player.Song
	if currentSong != nil {
		newPlaylist = append(newPlaylist, currentSong)
	}

	// Append non-nil songs to newPlaylist
	for _, song := range playlist {
		if song != nil {
			newPlaylist = append(newPlaylist, song)
		}
	}

	for i, song := range newPlaylist {
		if song == nil {
			continue
		}

		playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
		if i == 0 {
			playlistStr = fmt.Sprintf("%v <%v>\n\n", playlistStr, d.Player.GetCurrentStatus().String())
			embedMsg.SetThumbnail(song.Thumbnail.URL)
			if len(newPlaylist) > 1 {
				playlistStr += " **Next in queue:**\n"
			}
		}
	}

	embedMsg.SetDescription(playlistStr)
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
}
