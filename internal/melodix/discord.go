package melodix

import (
	"app/internal/config"
	"app/internal/version"
	"strconv"

	"fmt"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// BotInstance represents an instance of a Discord bot.
type BotInstance struct {
	Melodix *DiscordMelodix
}

// DiscordMelodix represents the Melodix instance for Discord.
type DiscordMelodix struct {
	Player         IMelodixPlayer
	Session        *discordgo.Session
	GuildID        string
	InstanceActive bool
	prefix         string
}

// NewDiscordMelodix creates a new instance of DiscordMelodix.
func NewDiscordMelodix(session *discordgo.Session, guildID string) *DiscordMelodix {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &DiscordMelodix{
		Player:         NewPlayer(guildID),
		Session:        session,
		InstanceActive: true,
		prefix:         config.DiscordCommandPrefix,
	}
}

// Start starts the DiscordMelodix instance.
func (dm *DiscordMelodix) Start(guildID string) {
	slog.Infof(`Discord instance started for guild id %v`, guildID)

	dm.Session.AddHandler(dm.Commands)
	dm.GuildID = guildID
}

// Commands handles incoming Discord commands.
func (dm *DiscordMelodix) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != dm.GuildID {
		return
	}

	if !dm.InstanceActive {
		return
	}

	command, parameter, err := parseCommand(m.Message.Content, dm.prefix)
	if err != nil {
		return
	}

	commandAliases := [][]string{
		{"pause", "!", ">"},
		{"resume", "play", ">"},
		{"play", "p", ">"},
		{"skip", "ff", ">>"},
		{"list", "queue", "l", "q"},
		{"add", "a", "+"},
		{"exit", "stop", "e", "x"},
		{"help", "h", "?"},
		{"history", "time", "t"},
		{"about", "v"},
	}

	canonicalCommand := getCanonicalCommand(command, commandAliases)
	if canonicalCommand == "" {
		return
	}

	switch canonicalCommand {
	case "pause":
		if parameter == "" && dm.Player.GetCurrentStatus() == StatusPlaying {
			dm.handlePauseCommand(s, m)
			return
		}
		fallthrough
	case "resume":
		if parameter == "" && dm.Player.GetCurrentStatus() != StatusPlaying {
			dm.handleResumeCommand(s, m)
			return
		}
		fallthrough
	case "play":
		dm.handlePlayCommand(s, m, parameter, false)
	case "skip":
		dm.handleSkipCommand(s, m)
	case "list":
		dm.handleShowQueueCommand(s, m)
	case "add":
		dm.handlePlayCommand(s, m, parameter, true)
	case "exit":
		dm.handleStopCommand(s, m)
	case "help":
		dm.handleHelpCommand(s, m)
	case "history":
		dm.handleHistoryCommand(s, m, parameter)
	case "about":
		dm.handleAboutCommand(s, m)
	default:
		// Unknown command
	}
}

// handlePlayCommand handles the play command for DiscordMelodix.
func (dm *DiscordMelodix) handlePlayCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string, enqueueOnly bool) {
	paramType, songsList := parseSongsAndTypeInParameter(param)

	if len(songsList) <= 0 {
		return
	}

	playlist := createPlaylist(paramType, songsList, dm, m)

	if len(playlist) > 0 {
		enqueuePlaylist(dm, playlist, s, m, enqueueOnly)
	} else {
		s.ChannelMessageSend(m.Message.ChannelID, "No songs to add to the queue.")
	}
}

// createPlaylist creates a playlist of songs based on the parameter type and list of songs.
func createPlaylist(paramType string, songsList []string, dm *DiscordMelodix, m *discordgo.MessageCreate) []*Song {
	var playlist []*Song

	for _, param := range songsList {
		var song *Song
		var err error
		switch paramType {
		case "id":
			id, err := strconv.Atoi(param)
			if err != nil {
				slog.Error("Cannot convert string id to int id")
				continue
			}
			song, err = FetchSongByID(m.GuildID, id)
			if err != nil {
				slog.Warnf("Error fetching song by history ID: %v", err)
			}
		case "title":
			song, err = FetchSongByTitle(param)
			if err != nil {
				slog.Warnf("Error fetching song by title: %v", err)
			}
		case "url":
			song, err = FetchSongByURL(param)
			if err != nil {
				slog.Warnf("Error fetching song by URL: %v", err)
			}
		}

		if song != nil {
			playlist = append(playlist, song)
		}
	}

	return playlist
}

// enqueuePlaylist enqueues a playlist of songs in the player's queue.
func enqueuePlaylist(dm *DiscordMelodix, playlist []*Song, s *discordgo.Session, m *discordgo.MessageCreate, enqueueOnly bool) {
	c, _ := s.State.Channel(m.Message.ChannelID)
	g, _ := s.State.Guild(c.GuildID)

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Message.Author.ID {
			if dm.Player.GetVoiceConnection() == nil {
				conn, err := dm.Session.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, true)
				if err != nil {
					slog.Errorf("Error connecting to voice channel: %v", err.Error())
					s.ChannelMessageSend(m.Message.ChannelID, "Error connecting to voice channel")
					return
				}
				dm.Player.SetVoiceConnection(conn)
				conn.LogLevel = discordgo.LogWarning
			}

			if len(playlist) > 0 {

				for _, song := range playlist {
					dm.Player.Enqueue(song)
				}

				embedMsg := embed.NewEmbed().
					SetColor(0x9f00d4).
					SetFooter(version.AppFullName)

				playlistStr := "🆕⁬ **Added to queue**\n\n"
				for i, song := range playlist {
					playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
				}

				embedMsg.SetDescription(playlistStr)
				message, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
				if err != nil {
					slog.Errorf("Error sending message: %v", err.Error())
					return
				}

				if !enqueueOnly && dm.Player.GetCurrentStatus() != StatusPlaying {
					go func() {
						for {
							if dm.Player.GetCurrentStatus() == StatusPlaying {

								embedMsg := embed.NewEmbed().
									SetColor(0x9f00d4).
									SetFooter(version.AppFullName)

								playlistStr := "▶️ **Playing**\n\n"
								for i, song := range playlist {
									playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
									if i == 0 {
										playlistStr = fmt.Sprintf("%v <%v>\n\n", playlistStr, dm.Player.GetCurrentStatus().String())
										embedMsg.SetThumbnail(song.Thumbnail.URL)
										if len(playlist) > 1 {
											playlistStr += " **Next in queue:**\n"
										}
									}
								}

								embedMsg.SetDescription(playlistStr)
								_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, message.ID, embedMsg.MessageEmbed)
								if err != nil {
									slog.Warnf("Error updating message: %v", err)
								}

								break
							}
						}
					}()
					dm.Player.Play(0, nil)
				}
			} else {
				s.ChannelMessageSend(m.Message.ChannelID, "No songs to add to the queue.")
			}
		}
	}
}

// handlePauseCommand handles the pause command for DiscordMelodix.
func (dm *DiscordMelodix) handlePauseCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	if dm.Player.GetCurrentSong().ID == "" {
		return
	}

	embedStr := "⏸ **Pause**"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	dm.Player.Pause()
}

// handleResumeCommand handles the resume command for DiscordMelodix.
func (dm *DiscordMelodix) handleResumeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	c, _ := s.State.Channel(m.Message.ChannelID)
	g, _ := s.State.Guild(c.GuildID)

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Message.Author.ID {
			if dm.Player.GetVoiceConnection() == nil {
				conn, err := dm.Session.ChannelVoiceJoin(c.GuildID, vs.ChannelID, false, true)
				if err != nil {
					slog.Errorf("Error connecting to voice channel: %v", err.Error())
					s.ChannelMessageSend(m.Message.ChannelID, "Error connecting to voice channel")
					return
				}
				dm.Player.SetVoiceConnection(conn)
				conn.LogLevel = discordgo.LogWarning
			}
		}
	}

	embedStr := "▶️ **Play (or resume)**"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	dm.Player.Unpause()
}

// handleStopCommand handles the stop command for DiscordMelodix.
func (dm *DiscordMelodix) handleStopCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	embedStr := "⏹ **Stop all activity**"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	dm.Player.ClearQueue()
	dm.Player.Stop()
}

// handleSkipCommand handles the skip command for DiscordMelodix.
func (dm *DiscordMelodix) handleSkipCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	embedStr := "⏩ **Skip track**"
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)

	dm.Player.Skip()
}

// handleShowQueueCommand handles the show queue command for DiscordMelodix.
func (dm *DiscordMelodix) handleShowQueueCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	embedMsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	// Get the current song and playlist
	currentSong := dm.Player.GetCurrentSong()
	playlist := dm.Player.GetSongQueue()

	// Check if there's a current song or the playlist is not empty
	if currentSong == nil && (len(playlist) == 0) {
		embedMsg.SetDescription("The queue is empty or no current song is playing.")
		s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
		return
	}

	playlistStr := "📑 **The queue**\n\n"

	var newPlaylist []*Song
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
			playlistStr = fmt.Sprintf("%v <%v>\n\n", playlistStr, dm.Player.GetCurrentStatus().String())
			embedMsg.SetThumbnail(song.Thumbnail.URL)
			if len(newPlaylist) > 1 {
				playlistStr += " **Next in queue:**\n"
			}
		}
	}

	embedMsg.SetDescription(playlistStr)
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
}

// handleHelpCommand handles the help command for DiscordMelodix.
func (dm *DiscordMelodix) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	play := fmt.Sprintf("**Play**: `%vplay [title/url/id]` \nAliases: `%vp [title/url/id]`, `%v> [title/url/id]`\n", dm.prefix, dm.prefix, dm.prefix)
	pause := fmt.Sprintf("**Pause** / **resume**: `%vpause`, `%vplay` \nAliases: `%v!`, `%v>`\n", dm.prefix, dm.prefix, dm.prefix, dm.prefix)
	queue := fmt.Sprintf("**Add track**: `%vadd [title/url/id]` \nAliases: `%va [title/url/id]`, `%v+ [title/url/id]`\n", dm.prefix, dm.prefix, dm.prefix)
	skip := fmt.Sprintf("**Skip track**: `%vskip` \nAliases: `%vff`, `%v>>`\n", dm.prefix, dm.prefix, dm.prefix)
	list := fmt.Sprintf("**Show queue**: `%vlist` \nAliases: `%vqueue`, `%vl`, `%vq`\n", dm.prefix, dm.prefix, dm.prefix, dm.prefix)
	history := fmt.Sprintf("**Show history**: `%vhistory`\n", dm.prefix)
	historyByDuration := fmt.Sprintf("**.. by duration**: `%vhistory duration`\n", dm.prefix)
	historyByPlaycount := fmt.Sprintf("**.. by play count**: `%vhistory count`\nAliases: `%vtime [count/duration]`, `%vt [count/duration]`", dm.prefix, dm.prefix, dm.prefix)
	stop := fmt.Sprintf("**Stop and exit**: `%vexit` \nAliases: `%ve`, `%vx`\n", dm.prefix, dm.prefix, dm.prefix)
	help := fmt.Sprintf("**Show help**: `%vhelp` \nAliases: `%vh`, `%v?`\n", dm.prefix, dm.prefix, dm.prefix)
	about := fmt.Sprintf("**Show version**: `%vabout`", dm.prefix)
	register := fmt.Sprintf("**Enable commands listening**: `%vregister`\n", dm.prefix)
	unregister := fmt.Sprintf("**Disable commands listening**: `%vunregister`", dm.prefix)

	embedMsg := embed.NewEmbed().
		SetTitle("ℹ️ Melodix — Command Usage").
		SetDescription("Some commands are aliased for shortness.\n`[title]` - track name\n`[url]` - youtube link\n`[id]` - track id from *History*.").
		AddField("", "*Playback*\n"+play+skip+pause).
		AddField("", "").
		AddField("", "*Queue*\n"+queue+list).
		AddField("", "").
		AddField("", "*History*\n"+history+historyByDuration+historyByPlaycount).
		AddField("", "").
		AddField("", "*General*\n"+stop+help+about).
		AddField("", "").
		AddField("", "*Administration*\n"+register+unregister).
		SetThumbnail("https://cdn.discordapp.com/app-icons/1137135371705122940/994ef64a83dd04d80c095efeb1dfdd2a.png?size=512").
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}

// handleHistoryCommand handles the history command for DiscordMelodix.
func (dm *DiscordMelodix) handleHistoryCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	var sortBy string
	var title string

	switch param {
	case "count", "times", "time":
		sortBy, title = "play_count", " — by play count"
	case "duration", "dur":
		sortBy, title = "duration", " — by total duration"
	default:
		sortBy, title = "last_played", " — most recent"
	}

	h := NewHistory()
	list, err := h.GetHistory(dm.GuildID, sortBy)
	if err != nil {
		slog.Warn("No history table found")
	}

	embedMsg := embed.NewEmbed().
		SetDescription(fmt.Sprintf("⏳ **History %v**", title)).
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	for _, elem := range list {
		duration := formatDuration(elem.History.Duration)
		fieldContent := fmt.Sprintf("```id: %d```    ```count: %d```    ```duration: %v```", elem.History.TrackID, elem.History.PlayCount, duration)

		embedMsg.AddField(fieldContent, fmt.Sprintf("[%v](%v)", elem.Track.Name, elem.Track.URL))
	}

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg.MessageEmbed)
}

// handleAboutCommand handles the about command for DiscordMelodix.
func (dm *DiscordMelodix) handleAboutCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	content := version.AppName + " is a simple music bot that allows you to play music in voice channels on a Discord server."

	embedStr := fmt.Sprintf("📻 **About %v**\n\n%v", version.AppName, content)
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		AddField("```"+version.BuildDate+"```", "Build date").
		AddField("```"+version.GoVersion+"```", "Go version").
		AddField("```Created by Innokentiy Sokolov```", "[Linkedin](https://www.linkedin.com/in/keshon), [GitHub](https://github.com/keshon), [Homepage](https://keshon.ru)").
		InlineAllFields().
		SetColor(0x9f00d4).SetFooter(version.AppFullName + " <" + dm.Player.GetCurrentStatus().String() + ">").MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}
