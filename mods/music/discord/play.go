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
	"github.com/keshon/melodix-player/mods/music/history"
	"github.com/keshon/melodix-player/mods/music/media"
	"github.com/keshon/melodix-player/mods/music/player"
	"github.com/keshon/melodix-player/mods/music/sources"
	"github.com/keshon/melodix-player/mods/music/utils"
)

func (d *Discord) handlePlayCommand(param string, enqueueOnly bool) {
	s := d.Session
	m := d.Message

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

	originType, origins := splitParamsToOriginsAndType(param)

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
	if d.Player.GetCurrentSong() != nil {
		enqueueOnly = true
	}
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

func getSongsFromSources(originType string, songsOrigins []string, guildID string) ([]*media.Song, error) {
	var songsList []*media.Song
	var allErrors []error // Slice to store all encountered errors

	youtube := sources.NewYoutube()
	stream := sources.NewStream()

	slog.Info("Getting songs list and their type:")
	for _, songOrigin := range songsOrigins {
		slog.Info(" - ", songOrigin, originType)
	}

	for _, songOrigin := range songsOrigins {

		var songs []*media.Song
		var err error

		switch originType {
		case "local_file":
			slog.Info("Local file: ", songOrigin)

			songPath := filepath.Join("cache", guildID, songOrigin)

			if _, err := os.Stat(songPath); os.IsNotExist(err) {
				allErrors = append(allErrors, fmt.Errorf("no such file or directory: %v", err))
				continue
			}

			var song *media.Song
			existingTrack, err := db.GetTrackByFilepath(songPath)
			if err == nil {
				song = &media.Song{
					SongID:   existingTrack.SongID,
					Title:    existingTrack.Title,
					URL:      existingTrack.URL,
					Filepath: existingTrack.Filepath,
				}
			} else {
				song = &media.Song{
					Title:    songOrigin,
					Filepath: songPath,
				}
			}

			song.Source = media.SourceLocalFile

			songs = append(songs, song)
		case "history_id":
			slog.Info("History ID: ", songOrigin)

			id, err := strconv.Atoi(songOrigin)
			if err != nil {
				allErrors = append(allErrors, fmt.Errorf("cannot convert string id to int id: %v", err))
				continue
			}
			h := history.NewHistory()
			track, err := h.GetTrackFromHistory(guildID, uint(id))
			if err != nil {
				slog.Error("Error getting track from history: %v", err)
				allErrors = append(allErrors, fmt.Errorf("error getting track from history with ID %v: %v", id, err))
				continue
			}

			var song []*media.Song

			switch track.Source {
			case "YouTube":
				slog.Info("Track is from YouTube")
				song, err = youtube.FetchManyByURL(track.URL)
				if err != nil {
					slog.Error("Error fetching song from youtube URL: %v", err)
					allErrors = append(allErrors, fmt.Errorf("%v", err))
				}
			case "Stream":
				slog.Info("Track is from Stream")
				song, err = stream.FetchManyByManyURLs([]string{track.URL})
				if err != nil {
					slog.Error("Error fetching stream from URL: %v", err)
					allErrors = append(allErrors, fmt.Errorf("%v", err))
				}
			case "LocalFile":
				slog.Info("Track is from LocalFile")
				song = []*media.Song{{
					SongID:   track.SongID,
					Title:    track.Title,
					URL:      track.URL,
					Filepath: track.Filepath,
					Source:   media.SourceLocalFile,
				}}
			}

			if allErrors != nil {
				allErrors = append(allErrors, fmt.Errorf("%v", err))
				continue
			}

			songs = append(songs, song...)
		case "youtube_title":
			slog.Info("Youtube title: ", songOrigin)
			songs, err = youtube.FetchManyByTitle(songOrigin)
			if err != nil {
				slog.Error("Error fetching song by title from youtube URL: %v", err)
				allErrors = append(allErrors, fmt.Errorf("%v", err))
			}
		case "youtube_url":
			slog.Info("Youtube URL: ", songOrigin)
			songs, err = youtube.FetchManyByManyURLs([]string{songOrigin})
			if err != nil {
				slog.Error("Error fetching song by URL from youtube URL: %v", err)
				allErrors = append(allErrors, fmt.Errorf("%v", err))
			}
		case "stream_url":
			slog.Info("Stream URL: ", songOrigin)
			songs, err = stream.FetchManyByManyURLs([]string{songOrigin})
			if err != nil {
				slog.Error("Error fetching stream from URL: %v", err)
				allErrors = append(allErrors, fmt.Errorf("%v", err))
			}
		}

		songsList = append(songsList, songs...)
	}

	if len(songsList) == 0 {
		return nil, fmt.Errorf("%v", allErrors)
	}

	slog.Info("Up-to-date songs list now is:")
	for _, song := range songsList {
		slog.Info(" - ", song.Title, song.Source)
	}

	return songsList, nil
}

func playOrEnqueue(d *Discord, playlist []*media.Song, s *discordgo.Session, m *discordgo.MessageCreate, enqueueOnly bool, prevMessageID string) (err error) {
	channel, err := s.State.Channel(m.Message.ChannelID)
	if err != nil {
		return err
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		return err
	}

	vs, found := d.findUserVoiceState(m.Message.Author.ID, guild.VoiceStates)
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
		playlist = d.Player.GetSongQueue()
		go func() {
			for {
				if d.Player.GetCurrentStatus() == player.StatusPlaying || d.Player.GetCurrentStatus() == player.StatusPaused {
					showStatusMessage(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist, false)
					break
				}
				time.Sleep(250 * time.Millisecond)
			}
		}()
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

		err := d.Player.Unpause(vs.ChannelID)
		if err != nil {
			return err
		}
	}

	slog.Warn("Current status is", d.Player.GetCurrentStatus().String())

	return nil
}

func showStatusMessage(d *Discord, s *discordgo.Session, channelID, prevMessageID string, playlist []*media.Song, previousPlaylistExist int, skipFirst bool) {
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4)

	playerStatus := fmt.Sprintf("%v %v", d.Player.GetCurrentStatus().StringEmoji(), d.Player.GetCurrentStatus().String())
	content := playerStatus + "\n"

	// Display current song information
	if currentSong := d.Player.GetCurrentSong(); currentSong != nil {
		var sourceLabels string
		if len(currentSong.URL) > 0 {

			if currentSong.Source == media.SourceLocalFile && utils.IsYouTubeURL(currentSong.URL) {
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
			content += fmt.Sprintf("\nNo song is currently playing, but the queue is filled with songs. Use `%vresume` command to toggle the playback\n\n", d.prefix)
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

func splitParamsToOriginsAndType(param string) (string, []string) {
	param = strings.TrimSpace(param)

	if len(param) == 0 {
		return "", []string{}
	}

	originType := ""

	// If has youtube keywords (assuming it's a link)
	if strings.Contains(param, "youtube") || strings.Contains(param, "youtu.be") {
		slog.Info("Possbile youtube urls: ", param)
		urlsSlice := strings.Fields(param)
		ytUrls := []string{}
		for _, url := range urlsSlice {
			if utils.IsYouTubeURL(url) {
				ytUrls = append(ytUrls, url)
				originType = "youtube_url"
			}
		}
		return originType, ytUrls
	} else {
		// Any non youtube link
		if strings.HasPrefix(param, "https") || strings.HasPrefix(param, "http") {
			slog.Info("Possible stream urls: ", param)
			urlsSlice := strings.Fields(param)
			streamUrls := []string{}
			for _, url := range urlsSlice {
				if utils.IsValidHttpURL(url) {
					streamUrls = append(streamUrls, url)
					originType = "stream_url"
				}
			}
			return originType, streamUrls
		} else {
			if strings.Contains(param, ".mp3") {
				slog.Info("Possible files: ", param)
				// Check if the parameter contains numeric IDs
				filesSlice := strings.Fields(param)
				files := []string{}
				for _, file := range filesSlice {
					if utils.IsAudioFile(file) {
						files = append(files, file)
						originType = "local_file"
					}
				}
				return originType, files
			} else {
				slog.Info("Possible history ids: ", param)
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

				// Threat as normal text (song title on youtube)
				// If none of the above conditions are met, treat it as a YouTube title
				encodedTitle := url.QueryEscape(param)
				return "youtube_title", []string{encodedTitle}
			}
		}
	}
}
