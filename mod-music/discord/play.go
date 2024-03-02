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
	"github.com/keshon/melodix-player/internal/version"
	"github.com/keshon/melodix-player/mod-music/player"
	"github.com/keshon/melodix-player/mod-music/sources"
)

// handlePlayCommand handles the play command for Discord.
func (d *Discord) handlePlayCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string, enqueueOnly bool) {
	d.changeAvatar(s)

	if param == "" {
		return
	}

	embedStr := "Please wait..."
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Error("Error sending 'please wait' message: %v", err)
	}

	originType, origins := parseOriginParameter(param)

	if len(origins) <= 0 {
		embedStr = "No songs or streams were found by your query."
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		if err != nil {
			slog.Error("Error sending 'please wait' message: %v", err)
		}
		return
	}

	channel, err := s.State.Channel(m.Message.ChannelID)
	if err != nil {
		slog.Error("Error getting channel: %v", err)
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		slog.Error("Error getting guild: %v", err)
	}

	if len(guild.VoiceStates) == 0 {
		embedStr = "You are not in a voice channel, please join one first."
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed
		_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		if err != nil {
			slog.Error("Error sending 'please wait' message: %v", err)
		}
		return
	}

	songs, err := fetchSongsToList(originType, origins, d, m)
	if err != nil {
		embedStr = fmt.Sprintf("%v\n\n*details:*\n`%v`", "Error forming playlist", err)
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed
		_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		if err != nil {
			slog.Error("Error sending 'please wait' message: %v", err)
		}
		return
	}

	// Enqueue playlist to the player
	err = playOrEnqueue(d, songs, s, m, enqueueOnly, pleaseWaitMessage.ID)
	if err != nil {
		slog.Error(err)

		embedStr = fmt.Sprintf("%v\n\n*details:*\n`%v`", "Error enqueuing/playing playlist", err)
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}
}

func fetchSongsToList(originType string, songsOrigins []string, d *Discord, m *discordgo.MessageCreate) ([]*player.Song, error) {
	var songsList []*player.Song

	youtube := sources.NewYoutube()
	stream := sources.NewStream()

	slog.Info("Getting songs list and their type:")
	for _, songOrigin := range songsOrigins {
		slog.Info(" - ", songOrigin, originType)
	}

	for _, songOrigin := range songsOrigins {

		var songs []*player.Song
		var err error

		switch originType {
		case "history_id":
			id, err := strconv.Atoi(songOrigin)
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
			songs, err = youtube.FetchSongsByTitle(songOrigin)
			if err != nil {
				slog.Warnf("Error fetching songs by title: %v", err)
				continue
			}
		case "youtube_url":
			songs, err = youtube.FetchSongsByURLs([]string{songOrigin})
			if err != nil {
				slog.Warnf("Error fetching songs by URL: %v", err)
				continue
			}
		case "stream_url":
			songs, err = stream.FetchStreamsByURLs([]string{songOrigin})
			if err != nil {
				slog.Warnf("Error fetching stream by URL: %v", err)
				continue
			}
		}

		songsList = append(songsList, songs...)
	}

	if len(songsList) == 0 {
		return nil, errors.New("no songs were fetched to playlist")
	}

	slog.Info("Up-to-date songs list now is:")
	for _, song := range songsList {
		slog.Info(" - ", song.Title, song.Source)
	}

	return songsList, nil
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
		d.Player.SetChannelID(vs.ChannelID)

		conn.LogLevel = discordgo.LogWarning
	}

	previousPlaylistExist := len(d.Player.GetSongQueue())

	// Enqueue songs
	slog.Info("Enqueuing the playlist to the player...")
	for _, song := range playlist {
		d.Player.Enqueue(song)
	}

	slog.Info("Player's song queue:")
	for _, song := range d.Player.GetSongQueue() {
		slog.Info(" - ", song.Title, song.Source)
	}

	if enqueueOnly {
		showStatusMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist, false)
		slog.Warn(d.Player.GetCurrentStatus().String())
	} else {
		go func() {
			for {
				if d.Player.GetCurrentStatus() == player.StatusPlaying || d.Player.GetCurrentStatus() == player.StatusPaused {
					showStatusMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist, true)
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
		}()

		slog.Warn("Current status is", d.Player.GetCurrentStatus().String())

		err := d.Player.Unpause(vs.ChannelID)
		if err != nil {
			return err
		}
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
			content += fmt.Sprintf("\nNo song is currently playing.\nUse the `%vplay [title/url/id/stream]` command to start. \nType `%vhelp` for more information.\n\n", d.prefix, d.prefix)
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

// parseOriginParameter parses the origin parameter and returns the appropriate type and value.
//
// param string - the parameter to be parsed
// (string, []string) - the type and value to be returned
func parseOriginParameter(param string) (string, []string) {
	param = strings.TrimSpace(param)

	if len(param) == 0 {
		return "", []string{}
	}

	u, err := url.Parse(param)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		paramSlice := strings.Fields(param)
		if u.Host == "www.youtube.com" || u.Host == "youtube.com" || u.Host == "youtu.be" {
			return "youtube_url", paramSlice
		}
		return "stream_url", paramSlice
	}

	params := strings.Fields(param)
	allValidIDs := true
	for _, id := range params {
		if _, err := strconv.Atoi(id); err != nil {
			allValidIDs = false
			break
		}
	}
	if allValidIDs {
		return "history_id", params
	}

	encodedTitle := url.QueryEscape(param)
	return "youtube_title", []string{encodedTitle}
}
