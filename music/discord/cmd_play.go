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

	// Wait message
	embedStr := getRandomWaitPhrase()
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)

	paramType, songsList := ParseSongsAndTypeInParameter(param)

	// Check if any songs were found
	if len(songsList) <= 0 {
		embedStr = "No music was found in request"
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
		embedStr = getVoiceChannelPhrase()
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

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
		embedStr = "No songs to queue"
		embedMsg = embed.NewEmbed().
			SetColor(0x9f00d4).
			SetDescription(embedStr).
			SetColor(0x9f00d4).MessageEmbed

		s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg)
		return
	}

	// Enqueue playlist to the player
	err = enqueuePlaylistV2(d, playlist, s, m, enqueueOnly, pleaseWaitMessage.ID)
	if err != nil {
		embedStr = fmt.Sprintf("Error enqueuing playlist: %v", err)
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

	previousPlaylistExist := len(d.Player.GetSongQueue())

	// Enqueue songs
	for _, song := range playlist {
		d.Player.Enqueue(song)
	}

	// Update playlist message
	if err := updatePlaylistMessage(s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist); err != nil {
		return err
	}

	// Start playing if not in enqueue-only mode
	if !enqueueOnly && d.Player.GetCurrentStatus() != player.StatusPlaying {
		go updatePlayingStatus(d, s, m.Message.ChannelID, prevMessageID, playlist, previousPlaylistExist)
		d.Player.Play(0, nil)
	}

	return nil
}

func updatePlaylistMessage(s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song, previousPlaylistExist int) error {
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
			if previousPlaylistExist > 0 {
				playlistContent = fmt.Sprintf("%v\n\n Some tracks have already been added — `!list` to see", playlistContent)
			}
		}
		playlistContent = fmt.Sprintf("%v\n` %v ` [%v](%v)", playlistContent, i, song.Name, song.UserURL)
	}

	embedMsg.SetDescription(playlistContent)

	_, err := s.ChannelMessageEditEmbed(channelID, prevMessageID, embedMsg.MessageEmbed)
	if err != nil {
		slog.Errorf("Error updating playlist message: %v", err)
		return err
	}

	return nil
}

func updatePlayingStatus(d *Discord, s *discordgo.Session, channelID, prevMessageID string, playlist []*player.Song, previousPlaylistExist int) {
	for {

		// Check if the player is in the playing status
		if d.Player.GetCurrentStatus() == player.StatusPlaying {
			embedMsg := embed.NewEmbed().
				SetColor(0x9f00d4).
				SetFooter(version.AppFullName)

			statusTitle := fmt.Sprintf("%v %v", d.Player.GetCurrentStatus().StringEmoji(), d.Player.GetCurrentStatus().String())
			nextTitle := "📑 In queue"
			playlistContent := statusTitle + "\n"

			for i, song := range playlist {

				if i == 0 {
					playlistContent = fmt.Sprintf("%v\n*[%v](%v)*", playlistContent, song.Name, song.UserURL)
					playlistContent = fmt.Sprintf("%v \n\n", playlistContent)
					embedMsg.SetThumbnail(song.Thumbnail.URL)
					if len(playlist) > 1 {
						playlistContent += nextTitle + "\n"
					}
				} else {
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
					playlistContent = fmt.Sprintf("%v\n` %v ` [%v](%v)", playlistContent, i, song.Name, song.UserURL)
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
		"Slow down, Captain Impatience!",
		"Hold your horses, I'm not Flash, ya know?",
		"Easy, I'm not racing a Formula 1 server here.",
		"I'm on it, just simmer down, okay?",
		"Take a breath, this isn't a comedy special.",
		"I move at the speed of a sloth on caffeine.",
		"Calm your coding cravings, I'm coding!",
		"Wait, you expected quantum speed? Cute.",
		"Relax, we're not launching rockets here.",
		"Your playlist is in line, like at the DMV.",
		"Hold tight, data's doing a stand-up routine.",
		"Patience, coding is not a fast-food drive-thru.",
		"I'm not a bot, I'm a chill algorithm.",
		"Put on your chill hat; we're taking our time.",
		"Hold up, servers need warm-up exercises.",
		"I'm not slow; I'm savoring the coding.",
		"Easy there, it's not a comedy roast server.",
		"Code's tap dancing—chill and enjoy it.",
		"Request's on a leisurely coding stroll.",
		"Hold on, I'm not a speedrun world record.",
		"Relax, servers are doing yoga poses.",
		"I code like I drive—cautious but steady.",
		"Your playlist is in the slow-cooker phase.",
		"Slow and steady, just like coding marathons.",
		"I'm not a sprinter; I'm a marathon coder.",
		"Your request's in line, like a patient cat.",
		"Coding's a dance, and we're waltzing.",
		"I'm not in a rush; I'm in a coding groove.",
		"Easy on the gas, we're not at Nascar.",
		"Chill vibes only; servers need a spa day.",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}
