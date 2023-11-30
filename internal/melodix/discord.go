package melodix

import (
	"app/internal/config"
	"app/internal/version"
	"strconv"
	"time"

	"fmt"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
)

// BotInstance represents an instance of a Discord bot.
type BotInstance struct {
	Melodix *Discord
}

// Discord represents the Melodix instance for Discord.
type Discord struct {
	Player               IPlayer
	Session              *discordgo.Session
	GuildID              string
	InstanceActive       bool
	prefix               string
	lastChangeAvatarTime time.Time
	rateLimitDuration    time.Duration
}

// NewDiscord creates a new instance of Discord.
func NewDiscord(session *discordgo.Session, guildID string) *Discord {
	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	return &Discord{
		Player:            NewPlayer(guildID),
		Session:           session,
		InstanceActive:    true,
		prefix:            config.DiscordCommandPrefix,
		rateLimitDuration: time.Minute * 10,
	}
}

// Start starts the Discord instance.
func (d *Discord) Start(guildID string) {
	slog.Infof(`Discord instance started for guild id %v`, guildID)

	d.Session.AddHandler(d.Commands)
	d.GuildID = guildID
}

// Commands handles incoming Discord commands.
func (d *Discord) Commands(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.GuildID != d.GuildID {
		return
	}

	if !d.InstanceActive {
		return
	}

	command, parameter, err := parseCommand(m.Message.Content, d.prefix)
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
		if parameter == "" && d.Player.GetCurrentStatus() == StatusPlaying {
			d.handlePauseCommand(s, m)
			return
		}
		fallthrough
	case "resume":
		if parameter == "" && d.Player.GetCurrentStatus() != StatusPlaying {
			d.handleResumeCommand(s, m)
			return
		}
		fallthrough
	case "play":
		d.handlePlayCommand(s, m, parameter, false)
	case "skip":
		d.handleSkipCommand(s, m)
	case "list":
		d.handleShowQueueCommand(s, m)
	case "add":
		d.handlePlayCommand(s, m, parameter, true)
	case "exit":
		d.handleStopCommand(s, m)
	case "help":
		d.handleHelpCommand(s, m)
	case "history":
		d.handleHistoryCommand(s, m, parameter)
	case "about":
		d.handleAboutCommand(s, m)
	default:
		// Unknown command
	}
}

// handlePlayCommand handles the play command for Discord.
func (d *Discord) handlePlayCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string, enqueueOnly bool) {
	d.changeAvatar(s)

	paramType, songsList := parseSongsAndTypeInParameter(param)

	if len(songsList) <= 0 {
		return
	}

	playlist, err := createPlaylist(paramType, songsList, d, m)
	if err != nil {
		s.ChannelMessageSend(m.Message.ChannelID, fmt.Sprintf("Error creating playlist: %v", err))
		return
	}

	if len(playlist) > 0 {
		enqueuePlaylist(d, playlist, s, m, enqueueOnly)
	} else {
		s.ChannelMessageSend(m.Message.ChannelID, "No songs to add to the queue.")
	}
}

// createPlaylist creates a playlist of songs based on the parameter type and list of songs.
func createPlaylist(paramType string, songsList []string, d *Discord, m *discordgo.MessageCreate) ([]*Song, error) {
	var playlist []*Song

	for _, param := range songsList {
		var songs []*Song
		var err error
		// var isManySongs bool
		switch paramType {
		case "id":
			id, err := strconv.Atoi(param)
			if err != nil {
				slog.Error("Cannot convert string id to int id")
				continue
			}
			songs, err = FetchSongsByID(m.GuildID, []int{id})
			if err != nil {
				slog.Warnf("Error fetching songs by history ID: %v", err)
			}
		case "title":
			songs, err = FetchSongsByTitle([]string{param})
			if err != nil {
				slog.Warnf("Error fetching songs by title: %v", err)
			}
		case "url":
			songs, err = FetchSongsByURL([]string{param})
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
func enqueuePlaylist(d *Discord, playlist []*Song, s *discordgo.Session, m *discordgo.MessageCreate, enqueueOnly bool) {
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

				embedsg := embed.NewEmbed().
					SetColor(0x9f00d4).
					SetFooter(version.AppFullName)

				playlistStr := "🆕⁬ **Added to queue**\n\n"
				for i, song := range playlist {
					playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
				}

				embedsg.SetDescription(playlistStr)
				message, err := s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg.MessageEmbed)
				if err != nil {
					slog.Errorf("Error sending message: %v", err.Error())
					return
				}

				if !enqueueOnly && d.Player.GetCurrentStatus() != StatusPlaying {
					go func() {
						for {
							if d.Player.GetCurrentStatus() == StatusPlaying {

								embedsg := embed.NewEmbed().
									SetColor(0x9f00d4).
									SetFooter(version.AppFullName)

								playlistStr := "▶️ **Playing**\n\n"
								for i, song := range playlist {
									playlistStr = fmt.Sprintf("%v%d. [%v](%v)\n", playlistStr, i+1, song.Name, song.UserURL)
									if i == 0 {
										playlistStr = fmt.Sprintf("%v <%v>\n\n", playlistStr, d.Player.GetCurrentStatus().String())
										embedsg.SetThumbnail(song.Thumbnail.URL)
										if len(playlist) > 1 {
											playlistStr += " **Next in queue:**\n"
										}
									}
								}

								embedsg.SetDescription(playlistStr)

								_, err := s.ChannelMessageEditEmbed(m.Message.ChannelID, message.ID, embedsg.MessageEmbed)
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

// handlePauseCommand handles the pause command for Discord.
func (d *Discord) handlePauseCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	if d.Player.GetCurrentSong().ID == "" {
		return
	}

	embedStr := "⏸ **Pause**"
	embedsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg)
	d.Player.Pause()
}

// handleResumeCommand handles the resume command for Discord.
func (d *Discord) handleResumeCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

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
		}
	}

	embedStr := "▶️ **Play (or resume)**"
	embedsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg)
	d.Player.Unpause()
}

// handleStopCommand handles the stop command for Discord.
func (d *Discord) handleStopCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	embedStr := "⏹ **Stop all activity**"
	embedsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg)
	d.Player.ClearQueue()
	d.Player.Stop()
}

// handleSkipCommand handles the skip command for Discord.
func (d *Discord) handleSkipCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	embedStr := "⏩ **Skip track**"
	embedsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg)

	d.Player.Skip()
}

// handleShowQueueCommand handles the show queue command for Discord.
func (d *Discord) handleShowQueueCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	embedsg := embed.NewEmbed().
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	// Get the current song and playlist
	currentSong := d.Player.GetCurrentSong()
	playlist := d.Player.GetSongQueue()

	// Check if there's a current song or the playlist is not empty
	if currentSong == nil && (len(playlist) == 0) {
		embedsg.SetDescription("The queue is empty or no current song is playing.")
		s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg.MessageEmbed)
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
			playlistStr = fmt.Sprintf("%v <%v>\n\n", playlistStr, d.Player.GetCurrentStatus().String())
			embedsg.SetThumbnail(song.Thumbnail.URL)
			if len(newPlaylist) > 1 {
				playlistStr += " **Next in queue:**\n"
			}
		}
	}

	embedsg.SetDescription(playlistStr)
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg.MessageEmbed)
}

// handleHelpCommand handles the help command for Discord.
func (d *Discord) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	play := fmt.Sprintf("**Play**: `%vplay [title/url/id]` \nAliases: `%vp [title/url/id]`, `%v> [title/url/id]`\n", d.prefix, d.prefix, d.prefix)
	pause := fmt.Sprintf("**Pause** / **resume**: `%vpause`, `%vplay` \nAliases: `%v!`, `%v>`\n", d.prefix, d.prefix, d.prefix, d.prefix)
	queue := fmt.Sprintf("**Add track**: `%vadd [title/url/id]` \nAliases: `%va [title/url/id]`, `%v+ [title/url/id]`\n", d.prefix, d.prefix, d.prefix)
	skip := fmt.Sprintf("**Skip track**: `%vskip` \nAliases: `%vff`, `%v>>`\n", d.prefix, d.prefix, d.prefix)
	list := fmt.Sprintf("**Show queue**: `%vlist` \nAliases: `%vqueue`, `%vl`, `%vq`\n", d.prefix, d.prefix, d.prefix, d.prefix)
	history := fmt.Sprintf("**Show history**: `%vhistory`\n", d.prefix)
	historyByDuration := fmt.Sprintf("**.. by duration**: `%vhistory duration`\n", d.prefix)
	historyByPlaycount := fmt.Sprintf("**.. by play count**: `%vhistory count`\nAliases: `%vtime [count/duration]`, `%vt [count/duration]`", d.prefix, d.prefix, d.prefix)
	stop := fmt.Sprintf("**Stop and exit**: `%vexit` \nAliases: `%ve`, `%vx`\n", d.prefix, d.prefix, d.prefix)
	help := fmt.Sprintf("**Show help**: `%vhelp` \nAliases: `%vh`, `%v?`\n", d.prefix, d.prefix, d.prefix)
	about := fmt.Sprintf("**Show version**: `%vabout`", d.prefix)
	register := fmt.Sprintf("**Enable commands listening**: `%vregister`\n", d.prefix)
	unregister := fmt.Sprintf("**Disable commands listening**: `%vunregister`", d.prefix)

	embedsg := embed.NewEmbed().
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
		AddField("", "*Adinistration*\n"+register+unregister).
		SetThumbnail("https://melodix-bot.keshon.ru/avatar/random"). // TODO: move out to config .env file
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg)
}

// handleHistoryCommand handles the history command for Discord.
func (d *Discord) handleHistoryCommand(s *discordgo.Session, m *discordgo.MessageCreate, param string) {
	d.changeAvatar(s)

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
	list, err := h.GetHistory(d.GuildID, sortBy)
	if err != nil {
		slog.Warn("No history table found")
	}

	embedsg := embed.NewEmbed().
		SetDescription(fmt.Sprintf("⏳ **History %v**", title)).
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName)

	for _, elem := range list {
		duration := formatDuration(elem.History.Duration)
		fieldContent := fmt.Sprintf("```id: %d```    ```count: %d```    ```duration: %v```", elem.History.TrackID, elem.History.PlayCount, duration)

		embedsg.AddField(fieldContent, fmt.Sprintf("[%v](%v)", elem.Track.Name, elem.Track.URL))
	}

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg.MessageEmbed)
}

// handleAboutCommand handles the about command for Discord.
func (d *Discord) handleAboutCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	content := version.AppName + " is a simple music bot that allows you to play music in voice channels on a Discord server."

	embedStr := fmt.Sprintf("📻 **About %v**\n\n%v", version.AppName, content)

	embedsg := embed.NewEmbed().
		SetDescription(embedStr).
		AddField("```"+version.BuildDate+"```", "Build date").
		AddField("```"+version.GoVersion+"```", "Go version").
		AddField("```Created by Innokentiy Sokolov```", "[Linkedin](https://www.linkedin.com/in/keshon), [GitHub](https://github.com/keshon), [Homepage](https://keshon.ru)").
		InlineAllFields().
		SetImage("https://melodix-bot.keshon.ru/avatar/random"). // TODO: move out to config .env file
		SetColor(0x9f00d4).SetFooter(version.AppFullName + " <" + d.Player.GetCurrentStatus().String() + ">").MessageEmbed

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedsg)
}

// changeAvatar changes bot avatar with randomly picked avatar image within allowed rate limit
func (d *Discord) changeAvatar(s *discordgo.Session) {
	// Check if the rate limit duration has passed since the last execution
	if time.Since(d.lastChangeAvatarTime) < d.rateLimitDuration {
		slog.Info("Rate-limited. Skipping changeAvatar.")
		return
	}

	imgPath, err := getRandomImagePathFromPath("./assets/avatars")
	if err != nil {
		slog.Errorf("Error getting avatar path: %v", err)
		return
	}

	avatar, err := readFileToBase64(imgPath)
	if err != nil {
		fmt.Printf("Error preparing avatar: %v\n", err)
		return
	}

	_, err = s.UserUpdate("", avatar)
	if err != nil {
		slog.Errorf("Error setting the avatar: %v", err)
		return
	}

	// Update the last execution time
	d.lastChangeAvatarTime = time.Now()
}

// createThumbnail returns encoded randomly picked avatar image
func createThumbnail() (string, error) {
	imgPath, err := getRandomImagePathFromPath("./assets/avatars")
	if err != nil {
		slog.Errorf("Error getting avatar path: %v", err)
		return "", err
	}

	avatar, err := readFileToBase64(imgPath)
	if err != nil {
		fmt.Printf("Error preparing avatar: %v\n", err)
		return "", err
	}

	return avatar, nil
}
