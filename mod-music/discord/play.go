package discord

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/mod-music/player"
	"github.com/keshon/melodix-discord-player/mod-music/sources"
)

// handlePlayCommand handles the play command for Discord.
func (d *Discord) handlePlayCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string, enqueueOnly bool) {
	d.changeAvatar(s)

	// If param actually has a value
	if param == "" {
		return
	}

	// Wait message
	embedStr := "Please wait..."
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Warnf("Error sending 'please wait' message: %v", err)
	}

	paramType, songsList := parseParameter(param)

	// Check if any songs were found
	if len(songsList) <= 0 {
		embedStr = "No songs or streams were found by your query."
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

	// Join voice channel message
	c, _ := s.State.Channel(m.Message.ChannelID)
	g, _ := s.State.Guild(c.GuildID)

	if len(g.VoiceStates) == 0 {
		embedStr = "You are not in a voice channel, please join one first."
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

	// Fill-in playlist
	playlist, err := createPlaylist(paramType, songsList, d, m)
	if err != nil {
		embedStr = fmt.Sprintf("%v\n\n*details:*\n`%v`", "Error forming playlist", err)
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed
		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

	if len(playlist) == 0 {
		embedStr = "There are no songs in the playlist."
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

	// Enqueue playlist to the player
	err = playOrEnqueue(d, playlist, s, m, enqueueOnly, pleaseWaitMessage.ID)
	if err != nil {
		embedStr = fmt.Sprintf("%v\n\n*details:*\n`%v`", "Error enqueuing/playing playlist", err)
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}
}

// createPlaylist creates a playlist of songs based on the parameter type and list of songs.
func createPlaylist(paramType string, songsList []string, d *Discord, m *discordgo.MessageCreate) ([]*player.Song, error) {
	var playlist []*player.Song

	youtube := sources.NewYoutube()
	stream := sources.NewStream()

	for _, param := range songsList {

		var songs []*player.Song
		var err error

		switch paramType {
		case "history_id":
			id, err := strconv.Atoi(param)
			if err != nil {
				slog.Error("Cannot convert string id to int id")
				continue
			}
			songs, err = youtube.FetchSongsByIDs(m.GuildID, []int{id})
			if err != nil {
				slog.Warnf("Error fetching songs by history ID: %v", err)
				continue
			}
		case "youtube_title":
			songs, err = youtube.FetchSongsByTitle(param)
			if err != nil {
				slog.Warnf("Error fetching songs by title: %v", err)
				continue
			}
		case "youtube_url":
			songs, err = youtube.FetchSongsByURLs([]string{param})
			if err != nil {
				slog.Warnf("Error fetching songs by URL: %v", err)
				continue
			}
		case "stream_url":
			songs, err = stream.FetchStreamsByURLs([]string{param})
			if err != nil {
				slog.Warnf("Error fetching stream by URL: %v", err)
				continue
			}
		}

		// if err != nil {
		// 	return nil, err
		// }

		playlist = append(playlist, songs...)
	}

	return playlist, nil
}

func playOrEnqueue(d *Discord, playlist []*player.Song, s *discordgo.Session, m *discordgo.MessageCreate, enqueueOnly bool, prevMessageID string) (err error) {
	channel, err := s.State.Channel(m.Message.ChannelID)
	if err != nil {
		return err
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return err
	}

	vs, found := findUserVoiceState(m.Message.Author.ID, guild.VoiceStates)
	if !found {
		return errors.New("user not found in voice channel")
	}

	if d.Player.GetVoiceConnection() == nil {
		conn, err := d.Session.ChannelVoiceJoin(channel.GuildID, vs.ChannelID, false, true)
		if err != nil {
			slog.Errorf("Error connecting to voice channel: %v", err.Error())
			s.ChannelMessageSend(m.Message.ChannelID, "Error connecting to voice channel")
			return err
		}
		d.Player.SetVoiceConnection(conn)
		conn.LogLevel = discordgo.LogWarning
	}

	previousPlaylistExist := len(d.Player.GetSongQueue())

	// Enqueue songs
	for _, song := range playlist {
		d.Player.Enqueue(song)
	}

	if enqueueOnly {
		showStatusMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist, false)
	} else {
		// Start goroutine to periodically check status
		go func() {
			for {
				if d.Player.GetCurrentStatus() == player.StatusPlaying || d.Player.GetCurrentStatus() == player.StatusPaused {
					showStatusMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist, true)
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
		}()

		d.Player.Play(0, nil)

		// // Start goroutine to initiate playback
		// go func() {
		// 	d.Player.Play(0, nil)
		// }()
		// for {
		// 	if d.Player.GetCurrentStatus() == player.StatusPlaying || d.Player.GetCurrentStatus() == player.StatusPaused {
		// 		showStatusMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist, true)
		// 		break
		// 	}
		// 	time.Sleep(250 * time.Millisecond)
		// }
	}

	return nil
}

func showStatusMessage(d *Discord, s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song, previousPlaylistExist int, skipFirst bool) {

	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	playerStatus := fmt.Sprintf("%v %v", d.Player.GetCurrentStatus().StringEmoji(), d.Player.GetCurrentStatus().String())
	content := playerStatus + "\n"

	// Display current song information
	if currentSong := d.Player.GetCurrentSong(); currentSong != nil {
		content += fmt.Sprintf("\n*[%v](%v)*\n\n", currentSong.Title, currentSong.UserURL)
		embedMsg.SetThumbnail(currentSong.Thumbnail.URL)
	} else {
		if len(d.Player.GetSongQueue()) > 0 {
			content += fmt.Sprintf("\nNo song is currently playing, but the queue is filled with songs. Use `%vplay` command to toggle the playback\n\n", d.prefix)
		} else {
			content += fmt.Sprintf("\nNo song is currently playing. Use the `%vplay [title/url/id/stream]` command to start. \nType `%vhelp` for more information.\n\n", d.prefix, d.prefix)
		}
	}

	// Display playlist information
	if len(playlist) > 0 {
		// Display queue status
		if !skipFirst || len(playlist) > 1 {
			content += "\n📑 In queue\n"
		}

		// Separate counter variable starting from 1
		counter := 1

		for i, song := range playlist {
			// Skip the first song if it's already playing
			if i == 0 && d.Player.GetCurrentSong() != nil && song == d.Player.GetCurrentSong() {
				continue
			}

			// Check if content length exceeds the limit
			if len(content) > 1800 {
				content = fmt.Sprintf("%v\n\nList too long to fit..", content)

				breakline := "\n"
				if previousPlaylistExist == 0 {
					breakline = "\n\n"
				}

				if previousPlaylistExist > 0 {
					content = fmt.Sprintf("%v%v Some tracks have already been added — `%vlist` to see", content, breakline, d.prefix)
				}
				break
			}

			// Display playlist entry
			content = fmt.Sprintf("%v\n` %v ` [%v](%v)", content, counter, song.Title, song.UserURL)
			counter++
		}
	}

	embedMsg.SetDescription(content)
	s.ChannelMessageEditEmbed(channelID, prevMessageID, embedMsg.MessageEmbed)
}

// ParseParameter parses the type and parameters from the input parameter string.
func parseParameter(param string) (string, []string) {
	// Trim spaces at the beginning and end
	param = strings.TrimSpace(param)

	if len(param) == 0 {
		return "", []string{}
	}

	// Check if the parameter is a URL
	u, err := url.Parse(param)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		// If it's a URL, split by ",", " ", new line, or carriage return
		paramSlice := strings.FieldsFunc(param, func(r rune) bool {
			return r == '\n' || r == '\r' || r == ' ' || r == '\t'
		})

		if isYouTubeURL(u.Host) {
			return "youtube_url", paramSlice
		} else {
			return "stream_url", paramSlice
		}
	}

	// Check if the parameter is an ID
	params := strings.Fields(param)
	allValidIDs := true
	for _, param := range params {
		_, err := strconv.Atoi(param)
		if err != nil {
			allValidIDs = false
			break
		}
	}
	if allValidIDs {
		return "history_id", params
	}

	// Treat it as a single title if it's not a URL or ID
	encodedTitle := url.QueryEscape(param)
	return "youtube_title", []string{encodedTitle}
}

// isYouTubeURL checks if the host is a YouTube URL.
func isYouTubeURL(host string) bool {
	return host == "www.youtube.com" || host == "youtube.com" || host == "youtu.be"
}
