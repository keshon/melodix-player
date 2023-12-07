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
	embedStr := getPleaseWaitPhrase()
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Warnf("Error sending 'please wait' message: %v", err)
	}

	paramType, songsList := ParseSongsAndTypeInParameter(param)

	// Check if any songs were found
	if len(songsList) <= 0 {
		embedStr = "I could not understand your song request"
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
	err = enqueuePlaylistV2(d, playlist, s, m, enqueueOnly, pleaseWaitMessage.ID)
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

func getErrorFormingPlaylistPhrase() string {
	phrases := []string{
		"Oopsie woopsie! Can't make a playlist, sowwy!",
		"Nyaa~ Sorry, playlist-making powers on cooldown!",
		"Teehee~ Playlist magic malfunction, my bad!",
		"UwU I tripped on code-kun! No playlist this time, sorry~",
		"Sowwy, playlist fairy got tangled in the code forest!",
		"Nyaa~ Playlist potion spilled! Apologies, senpai!",
		"Kawaii desu~ Playlist spell backfired! Sorry, sempai!",
		"UwU Playlist sprites are on vacation! Forgive me~",
		"Oops, playlist charm misfired! Forgive this game girl~",
		"Nyaa~ Playlist button is on a kawaii break, sorry!",
		"My bad, playlist's on a coffee break. Blame the intern.",
		"Oops, playlist chef had a stand-up gig. Sorry 'bout that.",
		"Playlist mixtape got lost in the comedy club. My bad.",
		"Sorry, playlist's on strike—demands more green M&Ms.",
		"Playlist DJ got caught in a Chappelle Show marathon. Oops.",
		"Playlist's playing hard to get, classic move. Forgive me.",
		"Playlist on a laughter yoga retreat. My apologies, friend.",
		"Playlist ghosted me. Even my code's getting swiped left.",
		"Oops, playlist's on vacation, sipping margaritas. My bad.",
		"Apologies, playlist's doing a comedy roast. Timing, right?",
		"Playlist machine took a day off. Classic.",
		"Playlist generator pulled a no-show. Go figure.",
		"Playlist system's on a spa day. Tough luck.",
		"Playlist magician called in sick. Surprise, surprise.",
		"Playlist computer is 'not feeling it today.' How novel.",
		"Playlist gizmo chose this moment to play hooky. Fantastic.",
		"Playlist contraption called in sick. Go figure, Ghandi.",
		"Playlist machine's MIA. Clearly, it's a genius move.",
		"Playlist sorcery is on strike. What a revelation, chief.",
		"Playlist rigamarole is ghosting us. Tough break, Snowflake.",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}

func getNoMusicFoundPhrase() string {
	phrases := []string{
		"No tunes matching your vibes.",
		"Music search came up empty.",
		"Can't find your requested beats.",
		"Sorry, no jams found.",
		"Your playlist is on a coffee break.",
		"No beats in this corner of the digital universe.",
		"Seems like the music elves are on vacation.",
		"Search yielded silence.",
		"No melody miracles today.",
		"Sorry, the sound waves went on strike.",
		"No music vibes detected.",
		"Playlist search ended up in a black hole.",
		"Beats MIA.",
		"404: Music not found.",
		"No harmonies in sight.",
		"Music radar malfunction.",
		"Looks like the songbird took a day off.",
		"Your beats are on vacation.",
		"Search party for your tunes canceled.",
		"No hits, just misses.",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}

func getJoinVoiceChannelPhrase() string {
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
		"You can't drop the beat in the void, join a voice channel!",
		"Can't spin the vinyl in empty space, join a voice channel first!",
		"Can't perform miracles in silence. Voice channel, please!",
		"Join the voice party, then let the beats drop...",
		"No voice channel, no music – it's science, man...",
		"Music needs an audience; voice channel, my friend...",
		"Can't groove in solitude; voice channel it up...",
		"Hold up! No voice channel, no sound waves...",
		"Join the voice channel; music awaits your ears...",
		"You're the missing link – get in a voice channel...",
		"No echoes in space; join a voice channel, genius...",
		"I'm a DJ, not a mind reader; voice channel first...",
		"No beats in the void; voice channel's the portal...",
		"Can't serenade the void; voice channel, my dude...",
		"Voiceless melodies? Join a channel, laugh track...",
		"Missing the voice memo? Channel up, amigo...",
		"Voiceless disco? Nah, join a channel, dance hero...",
		"Music in limbo? Nah, voice channel time, my friend...",
		"Beats on hold without voice; tune in, join up...",
		"Silent beats? Not here. Join a voice channel, boss...",
		"No beats in the void; voice channel's the link...",
		"Soundcheck's lonely; join a voice channel, maestro...",
		"Voiceless symphony? Join the channel orchestra...",
		"Can't spin airwaves; join a voice channel, champ...",
		"Voiceless DJ? Nah, join a channel, mix master...",
		"No voice, no beats; join a channel, music maestro...",
		"Join the chorus; voice channel's the VIP access...",
		"Can't serenade silence; voice channel, maestro...",
		"Beats on mute without voice; channel in, laugh out...",
		"Silent beats? Not on my watch. Voice channel, amigo...",
		"Voiceless gig? Join the channel comedy, my friend...",
		"Missing the voice memo? Channel up, dance down...",
		"No voice, no tunes; join a channel, groove king...",
		"Can't hum in vacuum; voice channel, genius move...",
		"Voiceless beats? Nah, join a channel, groove guru...",
		"No voice, no notes; join a channel, music wizard...",
		"Beats in exile without voice; channel in, laugh out...",
		"Silent disco? Not here. Voice channel, dance floor...",
		"No beats in the void; voice channel's the remedy...",
		"Voiceless melodies? Join a channel, laugh track...",
		"Can't groove in solitude; voice channel, dance party...",
		"Silent beats? Nah, join a channel, laugh out loud...",
		"No voice, no beats; join a channel, party starter...",
		"Can't serenade silence; voice channel, dance vibes...",
		"Beats on mute without voice; channel in, dance out...",
		"Silent disco? Not here. Voice channel, music magic...",
		"No voice, no notes; join a channel, laugh louder...",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}

func getPleaseWaitPhrase() string {
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
		"I'm on it, just simmer down, okay?",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}
