package discord

import (
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

	embedStr := GetVoiceChannelPhrase()
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

	embedStr = GetRandomWaitPhrase()
	embedMsg = embed.NewEmbed().
		SetColor(0x9f00d4).
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed

	pleaseWaitMessage, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Warnf("Error sending 'please wait' message: %v", err)
	}

	playlist, err := createPlaylist(paramType, songsList, d, m)
	if err != nil {
		s.ChannelMessageSend(m.Message.ChannelID, fmt.Sprintf("Error creating playlist: %v", err))
		return
	}

	if len(playlist) > 0 {
		enqueuePlaylist(d, playlist, s, m, enqueueOnly, pleaseWaitMessage)
	} else {
		s.ChannelMessageSend(m.Message.ChannelID, "No songs to add to the queue.")
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

// enqueuePlaylist enqueues a playlist of songs in the player's queue.
func enqueuePlaylist(d *Discord, playlist []*player.Song, s *discordgo.Session, m *discordgo.MessageCreate, enqueueOnly bool, pleaseWaitMessage *discordgo.Message) {
	c, _ := s.State.Channel(m.Message.ChannelID)
	g, _ := s.State.Guild(c.GuildID)

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Message.Author.ID {
			if d.Player.GetVoiceConnection() == nil {
				conn, err := d.Session.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, true)
				if err != nil {
					slog.Errorf("Error connecting to voice channel: %v", err.Error())
					s.ChannelMessageSend(m.Message.ChannelID, "Error connecting to voice channel")
					return
				}
				d.Player.SetVoiceConnection(conn)
				conn.LogLevel = discordgo.LogWarning
			}

			if len(playlist) > 0 {

				for _, song := range playlist {
					d.Player.Enqueue(song)
				}

				embedMsg := embed.NewEmbed().
					SetColor(0x9f00d4).
					SetFooter(version.AppFullName)

				playlistStr := "🆕⁬ **Added to queue**\n\n"
				for i, song := range playlist {
					playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
				}

				embedMsg.SetDescription(playlistStr)
				_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg.MessageEmbed)
				if err != nil {
					slog.Errorf("Error sending message: %v", err.Error())
					return
				}

				if !enqueueOnly && d.Player.GetCurrentStatus() != player.StatusPlaying {
					go func() {
						for {
							if d.Player.GetCurrentStatus() == player.StatusPlaying {

								embedMsg := embed.NewEmbed().
									SetColor(0x9f00d4).
									SetFooter(version.AppFullName)

								playlistStr := "▶️ **Playing**\n\n"
								for i, song := range playlist {
									playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
									if i == 0 {
										playlistStr = fmt.Sprintf("%v <%v>\n\n", playlistStr, d.Player.GetCurrentStatus().String())
										embedMsg.SetThumbnail(song.Thumbnail.URL)
										if len(playlist) > 1 {
											playlistStr += " **Next in queue:**\n"
										}
									}
								}

								embedMsg.SetDescription(playlistStr)

								_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, pleaseWaitMessage.ID, embedMsg.MessageEmbed)
								if err != nil {
									slog.Warnf("Error updating message: %v", err)
								}

								break
							}
						}
					}()
					d.Player.Play(0, nil)
				}
			} else {
				s.ChannelMessageSend(m.Message.ChannelID, "No songs to add to the queue.")
			}
		}
	}
}

func GetVoiceChannelPhrase() string {
	phrases := []string{
		"Hop into a voice channel, I'll be waiting...",
		"Can't serenade the silence, join a voice channel...",
		"Music needs an audience, join a voice channel first...",
		"Voice channels are where the party's at, join in...",
		"Can't play tunes in thin air, join a voice channel...",
		"Get ready to jam, but you gotta be in a voice channel...",
		"Time to turn up the volume, get into a voice channel...",
		"No silent disco here, join a voice channel for the beats...",
		"Let's make some noise, hop into a voice channel...",
		"Voice channels are the concert halls of Discord, join one...",
		"Music is meant to be heard, join a voice channel and let's roll...",
		"I'm ready to play, but you gotta be in a voice channel...",
		"Silent parties are no fun, join a voice channel and let's dance...",
		"Get your vocal cords ready, join a voice channel for the tunes...",
		"Step into the arena of sound, join a voice channel now...",
		"I can't play music in thin air, join a voice channel first...",
		"Turn on the mic, turn up the music, join a voice channel...",
		"Grab a seat in the virtual amphitheater, join a voice channel...",
		"Join the audio adventure, get into a voice channel...",
		"Where words fail, music speaks – join a voice channel...",
		"The party is where the voices are, join a channel to start...",
		"Unlock the music vault by entering a voice channel...",
		"Can't serenade empty spaces, join a voice channel...",
		"Voice channels are the DJ booths of Discord, join the party...",
		"Music's calling, but it needs a voice channel to answer...",
		"Prepare for an audio journey, hop into a voice channel...",
		"Music's knocking on your eardrums, open the door with a voice channel...",
		"Ready to broadcast tunes, but you gotta be in a voice channel...",
		"Silence isn't golden when it comes to music – join a voice channel...",
		"Get cozy in a voice channel, the music's about to begin...",
		"Unlock the melody, step into a voice channel...",
		"Join the sonic revolution – get into a voice channel...",
		"The stage is set, but you need to step into a voice channel...",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}

func GetRandomWaitPhrase() string {
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
