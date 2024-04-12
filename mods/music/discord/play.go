package discord

import (
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/db"
	"github.com/keshon/melodix-player/internal/version"
	"github.com/keshon/melodix-player/mods/music/history"
	"github.com/keshon/melodix-player/mods/music/player"
	"github.com/keshon/melodix-player/mods/music/sources"
	"github.com/keshon/melodix-player/mods/music/utils"
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

	songs, err := getSongsFromSources(originType, origins, m.GuildID)
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

func getSongsFromSources(originType string, songsOrigins []string, guildID string) ([]*player.Song, error) {
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
		case "local_file":
			slog.Info("Local file: ", songOrigin)

			songPath := filepath.Join("cache", guildID, songOrigin)

			if _, err := os.Stat(songPath); os.IsNotExist(err) {
				slog.Error("No such file or directory: %v", err)
				continue
			}

			var song *player.Song
			existingTrack, err := db.GetTrackByFilepath(songPath)
			if err == nil {
				song = &player.Song{
					SongID:   existingTrack.SongID,
					Title:    existingTrack.Title,
					URL:      existingTrack.URL,
					Filepath: existingTrack.Filepath,
				}
			} else {
				song = &player.Song{
					Title:    songOrigin,
					Filepath: songPath,
				}
			}

			song.Source = player.SourceLocalFile

			songs = append(songs, song)
		case "history_id":
			slog.Info("History ID: ", songOrigin)

			id, err := strconv.Atoi(songOrigin)
			if err != nil {
				slog.Error("Cannot convert string id to int id")
				continue
			}
			h := history.NewHistory()
			track, err := h.GetTrackFromHistory(guildID, uint(id))
			slog.Error(track)
			if err != nil {
				slog.Error("Error getting track from history with ID %v", id)
				continue
			}

			var song []*player.Song

			if track.Source == "YouTube" {
				slog.Info("Track is from YouTube")
				song, err = youtube.GetAllSongsFromURL(track.URL)
				if err != nil {
					slog.Error("error fetching new songs from URL: %v", err)
					continue
				}
			}
			if track.Source == "Stream" {
				slog.Info("Track is from Stream")
				song, err = stream.FetchStreamsByURLs([]string{track.URL})
				if err != nil {
					slog.Error("error fetching new songs from URL: %v", err)
					continue
				}
			}
			if track.Source == "LocalFile" {
				slog.Info("Track is from LocalFile")
				song = []*player.Song{{
					SongID:   track.SongID,
					Title:    track.Title,
					URL:      track.URL,
					Filepath: track.Filepath,
					Source:   player.SourceLocalFile,
				}}
			}

			songs = append(songs, song...)
		case "youtube_title":
			slog.Info("Youtube title: ", songOrigin)

			songs, err = youtube.FetchSongsByTitle(songOrigin)
			if err != nil {
				slog.Warnf("Error fetching songs by title: %v", err)
				continue
			}
		case "youtube_url":
			slog.Info("Youtube URL: ", songOrigin)

			songs, err = youtube.FetchSongsByURLs([]string{songOrigin})
			if err != nil {
				slog.Warnf("Error fetching songs by URL: %v", err)
				continue
			}
		case "stream_url":
			slog.Info("Stream URL: ", songOrigin)

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
		var sourceLabels string
		if len(currentSong.URL) > 0 {

			if currentSong.Source == player.SourceLocalFile && utils.IsYouTubeURL(currentSong.URL) {
				sourceLabels = "`youtube cached`"
			} else {
				sourceLabels = "`" + strings.ToLower(currentSong.Source.String()) + "`"
			}

			content += fmt.Sprintf("\n**`%v`**\n[%v](%v)\n\n", sourceLabels, currentSong.Title, currentSong.URL)
		} else {
			content += fmt.Sprintf("\n**`%v`**\n%v\n\n", strings.ToLower(currentSong.Source.String()), currentSong.Title)
		}
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

		for i, elem := range playlist {
			// Skip the first song if it's already playing
			if i == 0 && d.Player.GetCurrentSong() != nil && elem == d.Player.GetCurrentSong() {
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
			if elem.URL != "" {
				content = fmt.Sprintf("%v\n` %v ` [%v](%v)", content, counter, elem.Title, elem.URL)
			} else {
				content = fmt.Sprintf("%v\n` %v ` %v", content, counter, elem.Title)
			}
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

	// Check if the parameter is a URL
	u, err := url.Parse(param)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") {
		paramSlice := strings.Fields(param)
		if utils.IsYouTubeURL(u.Host) {
			return "youtube_url", paramSlice
		}
		return "stream_url", paramSlice
	}

	// Check if the parameter contains numeric IDs
	allValidIDs := true
	for _, id := range strings.Fields(param) {
		if _, err := strconv.Atoi(id); err != nil {
			allValidIDs = false
			break
		}
	}
	if allValidIDs {
		return "history_id", strings.Fields(param)
	}

	// Check if the parameter contains file extensions
	for _, part := range strings.Fields(param) {
		if strings.HasSuffix(part, ".ac3") || strings.HasSuffix(part, ".aac") || strings.HasSuffix(part, ".opus") || strings.HasSuffix(part, ".mp3") || strings.HasSuffix(part, ".m4a") {
			return "local_file", []string{part}
		}
	}

	// If none of the above conditions are met, treat it as a YouTube title
	encodedTitle := url.QueryEscape(param)
	return "youtube_title", []string{encodedTitle}
}
