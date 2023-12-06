package discord

import (
	"errors"
	"fmt"
	"math/rand"
	"strconv"

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

	paramType, songsList := ParseSongsAndTypeInParameter(param)

	if len(songsList) <= 0 {
		return
	}

	// Join voice channel message
	embedStr := getVoiceChannelPhrase()
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	c, _ := s.State.Channel(m.Message.ChannelID)
	g, _ := s.State.Guild(c.GuildID)

	if len(g.VoiceStates) == 0 {
		s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
		return
	}

	// Wait message
	embedStr = getRandomWaitPhrase()
	embedMsg = embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Warnf("Error sending 'please wait' message: %v", err)
	}

	// Fill-in playlist
	playlist, err := createPlaylist(paramType, songsList, d, m)
	if err != nil {
		s.ChannelMessageSend(m.Message.ChannelID, fmt.Sprintf("Error creating playlist: %v", err))
		return
	}

	if len(playlist) == 0 {
		s.ChannelMessageSend(m.Message.ChannelID, "No songs to enqueue.")
		return
	}

	// Enqueue playlist to the player
	err = enqueuePlaylistV2(d, playlist, s, m, enqueueOnly, pleaseWaitMessage.ID)
	if err != nil {
		s.ChannelMessageSend(m.Message.ChannelID, fmt.Sprintf("Error enqueuing playlist: %v", err))
		return
	}
}

// createPlaylist creates a playlist of songs based on the parameter type and list of songs.
func createPlaylist(paramType string, songsList []string, d *Discord, m *discordgo.MessageCreate) ([]*player.Song, error) {
	var playlist []*player.Song

	youtube := sources.NewYoutube()
	for _, param := range songsList {
		var songs []*player.Song
		var err error
		// var isManySongs bool
		switch paramType {
		case "id":
			id, err := strconv.Atoi(param)
			if err != nil {
				slog.Error("Cannot convert string id to int id")
				continue
			}
			songs, err = youtube.FetchSongsByIDs(m.GuildID, []int{id})
			if err != nil {
				slog.Warnf("Error fetching songs by history ID: %v", err)
			}
		case "title":
			songs, err = youtube.FetchSongsByTitle(param)
			if err != nil {
				slog.Warnf("Error fetching songs by title: %v", err)
			}
		case "url":
			songs, err = youtube.FetchSongsByURLs([]string{param})
			if err != nil {
				slog.Warnf("Error fetching songs by URL: %v", err)
			}
		}

		if err != nil {
			return nil, err
		}

		playlist = append(playlist, songs...)
	}

	return playlist, nil
}

func enqueuePlaylistV2(d *Discord, playlist []*player.Song, s *discordgo.Session, m *discordgo.MessageCreate, enqueueOnly bool, prevMessageID string) (err error) {
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

	// Enqueue songs
	for _, song := range playlist {
		d.Player.Enqueue(song)
	}

	// Update playlist message
	if err := updatePlaylistMessage(s, m.Message.ChannelID, prevMessageID, playlist); err != nil {
		return err
	}

	// Start playing if not in enqueue-only mode
	if !enqueueOnly && d.Player.GetCurrentStatus() != player.StatusPlaying {
		go updatePlayingStatus(d, s, m.Message.ChannelID, prevMessageID, playlist)
		d.Player.Play(0, nil)
	}

	return nil
}

func updatePlaylistMessage(s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song) error {
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	playlistStr := "🆕 **Added to queue**\n\n"
	for _, song := range playlist {
		playlistStr = fmt.Sprintf("%v- [%v](%v)\n", playlistStr, song.Name, song.UserURL)
	}

	embedMsg.SetDescription(playlistStr)

	_, err := s.ChannelMessageEditEmbed(channelID, prevMessageID, embedMsg.MessageEmbed)
	if err != nil {
		slog.Errorf("Error updating playlist message: %v", err)
		return err
	}

	return nil
}

func updatePlayingStatus(d *Discord, s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song) {
	for {
		// Check if the player is in the playing status
		if d.Player.GetCurrentStatus() == player.StatusPlaying {
			embedMsg := embed.NewEmbed().
				SetColor(0x9f00d4).
				SetFooter(version.AppFullName)

			statusTitle := fmt.Sprintf("%v %v", d.Player.GetCurrentStatus().StringEmoji(), d.Player.GetCurrentStatus().String())
			nextTitle := "📑 Next"
			playlistContent := statusTitle + "\n"

			for i, song := range playlist {
				playlistContent = fmt.Sprintf("%v- [%v](%v)\n", playlistContent, song.Name, song.UserURL)
				if i == 0 {
					playlistContent = fmt.Sprintf("%v \n\n", playlistContent)
					embedMsg.SetThumbnail(song.Thumbnail.URL)
					if len(playlist) > 1 {
						playlistContent += nextTitle + "\n"
					}
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

func getVoiceChannelPhrase() string {
	phrases := []string{
		"Hop into a voice channel, then try again...",
		"Can't serenade the silence, join a voice channel first...",
		"Music needs an audience, join a voice channel first...",
		"Can't play tunes in thin air, join a voice channel...",
		"You gotta be in a voice channel...",
		"Get into a voice channel...",
		"No silent disco here, join a voice channel first...",
		"Hop into a voice channel first...",
		"Music is meant to be heard, join a voice channel first...",
		"You gotta be in a voice channel...",
		"I can't play music in thin air, join a voice channel first...",
		"Can't serenade empty spaces, join a voice channel first...",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}

func getRandomWaitPhrase() string {
	phrases := []string{
		"Chillax, I'm on it...",
		"Easy there, turbo...",
		"Ever heard of fashionably late?",
		"Hold your horses, we got this...",
		"Patience, my young padawan...",
		"I move at my own pace, deal with it...",
		"Slow and steady wins the race, right?",
		"Taking my time, just like a fine wine...",
		"Can't rush perfection, my friend...",
		"Grab a snack, this might take a minute...",
		"Tick-tock, but in my own clock...",
		"Did someone order a chilled response?",
		"Sit back, relax, and enjoy the show...",
		"Don't rush me, I'm on island time...",
		"Mastering the art of fashionably late...",
		"Patience, grasshopper...",
		"Hang in there, superstar...",
		"Hold my server, I got this...",
		"Data's doing the cha-cha...",
		"Server's got moves, wait...",
		"Code's flexing its muscles...",
		"Binary bits breakdancing...",
		"Servers tap dancing for you...",
		"Coding wizardry in progress...",
		"Request on a magic carpet...",
		"Cyber monkeys typing furiously...",
		"Your wish is my command...almost...",
		"Quantum computing, almost there...",
		"Data sprinting to your screen...",
		"Virtual acrobatics in motion...",
		"Code juggling like a boss...",
		"Bytes breakdancing in the server...",
		"Request breakdancing through firewalls...",
		"Code tap dancing its way...",
		"Server's telling knock-knock jokes...",
		"Request on a virtual rollercoaster...",
		"Algorithms breakdancing for you...",
		"Ninja moves on your request...",
		"Coffee break while we work...",
		"Request moonwalking to completion...",
		"Wild times in the server room...",
		"Sit back, enjoy the show...",
		"Sloth could be faster, but we're on it...",
		"Grab popcorn, it's interesting...",
		"Your request is the VIP...",
		"Put on a seatbelt, bumpy ride...",
		"Request on a data rollercoaster...",
		"Cha-cha with our servers...",
		"Counting to infinity... almost done...",
		"Brace yourself, request is dropping...",
		"Working harder than a cat...",
		"Fairy dust, request complete...",
		"Hold on tight, breakdancing to you...",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}
