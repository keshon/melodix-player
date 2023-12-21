package discord

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/music/player"
	"github.com/keshon/melodix-discord-player/music/sources"
)

// handlePlayCommand handles the play command for Discord.
func (d *Discord) handlePlayCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string, enqueueOnly bool) {
	d.changeAvatar(s)

	// Wait message
	embedStr := getPleaseWaitPhrase()
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
		embedStr = getErrorRequestPhrase()
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
		embedStr = getJoinVoiceChannelPhrase()
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
		embedStr = fmt.Sprintf("%v\n\n**Error details**:\n`%v`", getErrorFormingPlaylistPhrase(), err)
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed
		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

	if len(playlist) == 0 {
		embedStr = getNoMusicFoundPhrase()
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
		embedStr = fmt.Sprintf("%v\n\n**Error details**:\n`%v`", getErrorFormingPlaylistPhrase(), err)
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
		// var isManySongs bool
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
			}
		case "youtube_title":
			songs, err = youtube.FetchSongsByTitle(param)
			if err != nil {
				slog.Warnf("Error fetching songs by title: %v", err)
			}
		case "youtube_url":
			songs, err = youtube.FetchSongsByURLs([]string{param})
			if err != nil {
				slog.Warnf("Error fetching songs by URL: %v", err)
			}
		case "stream_url":
			// TODO: implement adding radio stations URLs
			songs, err = stream.FetchStreamsByURLs([]string{param})
		}

		if err != nil {
			return nil, err
		}

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

	// Update playlist message
	if err := updateAddToQueueMessage(s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist); err != nil {
		return err
	}

	// Start playing if not in enqueue-only mode
	if !enqueueOnly && d.Player.GetCurrentStatus() != player.StatusPlaying {
		go updatePlayingNowMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist)
		d.Player.Play(0, nil)
	}

	return nil
}

func updateAddToQueueMessage(s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song, previousPlaylistExist int) error {
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	nextTitle := "📑 In queue"
	playlistContent := nextTitle + "\n"

	for i, song := range playlist {
		if len(playlistContent) > 1800 {
			playlistContent = fmt.Sprintf("%v\n\nList too long to fit..", playlistContent)

			breakline := "\n"
			if previousPlaylistExist == 0 {
				breakline = "\n\n"
			}
			if previousPlaylistExist > 0 {
				playlistContent = fmt.Sprintf("%v%v Some tracks have already been added — `!list` to see", playlistContent, breakline)
			}
			break
		} else if i == len(playlist)-1 {
			playlistContent = fmt.Sprintf("%v\n` %v ` [%v](%v)", playlistContent, i, song.Title, song.UserURL)
			if previousPlaylistExist > 0 {
				playlistContent = fmt.Sprintf("%v\n\n Some tracks have already been added — `!list` to see", playlistContent)
			}
		} else {
			playlistContent = fmt.Sprintf("%v\n` %v ` [%v](%v)", playlistContent, i, song.Title, song.UserURL)
		}

	}

	embedMsg.SetDescription(playlistContent)

	_, err := s.ChannelMessageEditEmbed(channelID, prevMessageID, embedMsg.MessageEmbed)
	if err != nil {
		slog.Errorf("Error updating playlist message: %v", err)
		return err
	}

	return nil
}

func updatePlayingNowMessage(d *Discord, s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song, previousPlaylistExist int) {
	for {

		// Check if the player is in the playing status
		if d.Player.GetCurrentStatus() == player.StatusPlaying {
			embedMsg := embed.NewEmbed().
				SetColor(0x9f00d4).
				SetFooter(version.AppFullName)

			statusTitle := fmt.Sprintf("%v %v", d.Player.GetCurrentStatus().StringEmoji(), d.Player.GetCurrentStatus().String())
			slog.Info(statusTitle)
			nextTitle := "📑 In queue"
			var playlistContent string

			// if d.Player.GetCurrentSong() != nil {
			playlistContent = statusTitle + "\n"
			playlistContent = fmt.Sprintf("%v\n*[%v](%v)*\n\n", playlistContent, d.Player.GetCurrentSong().Title, d.Player.GetCurrentSong().UserURL)
			embedMsg.SetThumbnail(d.Player.GetCurrentSong().Thumbnail.URL)
			// }

			if len(playlist) > 1 {
				playlistContent += nextTitle + "\n"

				// Separate counter variable starting from 1
				counter := 1

				for i, song := range playlist {
					if i == 0 {
						if song == d.Player.GetCurrentSong() {
							// Skip the first song if it's already playing
							continue
						}
					}

					if len(playlistContent) > 1800 {
						playlistContent = fmt.Sprintf("%v\n\nList too long to fit..", playlistContent)

						breakline := "\n"
						if previousPlaylistExist == 0 {
							breakline = "\n\n"
						}
						if previousPlaylistExist > 0 {
							playlistContent = fmt.Sprintf("%v%v Some tracks have already been added — `%vlist` to see", playlistContent, breakline, d.prefix)
						}
						break
					} else if i == len(playlist)-1 {
						if previousPlaylistExist > 0 {
							playlistContent = fmt.Sprintf("%v\n\n Some tracks have already been added — `%vlist` to see", playlistContent, d.prefix)
						}
					}

					// Use the separate counter variable for display
					playlistContent = fmt.Sprintf("%v\n` %v ` [%v](%v)", playlistContent, counter, song.Title, song.UserURL)

					// Increment the counter for each iteration
					counter++
				}
			}

			embedMsg.SetDescription(playlistContent)

			_, err := s.ChannelMessageEditEmbed(channelID, prevMessageID, embedMsg.MessageEmbed)
			if err != nil {
				slog.Warnf("Error updating playing status message: %v", err)
			}

			break
		}
	}
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
