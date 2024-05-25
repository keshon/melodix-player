package discord

import (
	"fmt"
	"os"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/internal/version"
	"github.com/keshon/melodix-player/mods/about/utils"
)

func (d *Discord) handleHelpCommand(param string) {
	s := d.Session
	m := d.Message
	d.changeAvatar(s)

	cfg, err := config.NewConfig()
	if err != nil {
		slog.Fatal("Error loading config:", err)
		return
	}

	host := os.Getenv("HOST")
	if host == "" {
		host = cfg.RestHostname
	}

	avatarURL := utils.InferProtocolByPort(host, 443) + host + "/avatar/random?" + fmt.Sprint(time.Now().UnixNano())

	switch param {
	case "play":
		d.handleHelpPlay()
		return
	case "queue":
		d.handleHelpQueue()
		return
	case "history":
		d.handleHelpHistory()
		return
	case "info":
		d.handleHelpInfo()
		return
	case "manage":
		d.handleHelpManage()
		return
	case "cache":
		d.handleHelpCache()
		return
	}

	prefix := d.CommandPrefix

	play := fmt.Sprintf("`%vplay [title|url|stream|id]` — play selected track/radio\n", prefix)
	skip := fmt.Sprintf("`%vskip` — play next track\n", prefix)
	pause := fmt.Sprintf("`%vpause`, `%vresume` — pause/resume playback\n", prefix, prefix)
	stop := fmt.Sprintf("`%vstop` — stop playback and leave voice channel\n", prefix)

	add := fmt.Sprintf("`%vadd [title/url/id]` — add track\n", prefix)
	list := fmt.Sprintf("`%vlist` — show current queue\n", prefix)

	history := fmt.Sprintf("`%vhistory` — show played tracks\n", prefix)
	historyByDuration := fmt.Sprintf("`%vhistory duration` — sort by duration \n", prefix)
	historyByPlaycount := fmt.Sprintf("`%vhistory count` — sort by play count \n", prefix)

	cached := fmt.Sprintf("`%vcached` — show cached tracks\n", prefix)
	cachedSync := fmt.Sprintf("`%vcached sync` — sync added/removed files to with database\n", prefix)
	curl := fmt.Sprintf("`%vcurl [url]` — cache track (youtube url only)\n", prefix)
	uploaded := fmt.Sprintf("`%vuploaded` — show uploaded videos\n", prefix)
	uploadedExtract := fmt.Sprintf("`%vuploaded extract` — extract audio from uploaded videos to cache\n", prefix)

	help := fmt.Sprintf("`%vhelp`, `%vh` — show help\n", prefix, prefix)
	about := fmt.Sprintf("`%vabout`, `%vv` — show version\n", prefix, prefix)
	whoami := fmt.Sprintf("`%vwhoami` — log user's info\n", prefix)

	register := fmt.Sprintf("`%vregister` — enable commands listening\n", prefix)
	unregister := fmt.Sprintf("`%vunregister` — disable commands listening\n", prefix)
	melodixPrefix := "`melodix-prefix` — print current command prefix\n"
	melodixPrefixUpdate := "`melodix-prefix-update \"[new_prefix]\"` — set new prefix (in quotes)\n"
	melodixPreifxReset := fmt.Sprintf("`melodix-prefix-reset` — reset prefix to global one: `%v`\n", cfg.DiscordCommandPrefix)

	title := fmt.Sprintf("ℹ️ %v — Commands Usage\n\n", version.AppName)

	embedMsg := embed.NewEmbed().
		SetDescription(title+"[title] - track name\n[url] - YouTube URL\n[id] - track id from *History*\n[stream] - valid stream URL (radio).\n\n").
		AddField("", "**Playback**\n"+play+skip+pause+stop+"\n`"+prefix+"help play` for more..\n").
		AddField("", "").
		AddField("", "**Queue**\n"+add+list+"\n`"+prefix+"help queue` for more..\n").
		AddField("", "").
		AddField("", "**History**\n"+history+historyByDuration+historyByPlaycount+"\n`"+prefix+"help history` for more..\n").
		AddField("", "").
		AddField("", "**Information**"+help+about+whoami+"\n`"+prefix+"help info` for more..\n").
		AddField("", "").
		AddField("", "**Management**"+register+unregister+melodixPrefix+melodixPrefixUpdate+melodixPreifxReset+"\n`"+prefix+"help manage` for more..\n").
		AddField("", "").
		AddField("", "**Caching & Sideloading**\nThis commands are for superadmin only.\n"+cached+cachedSync+curl+uploaded+uploadedExtract+"\n`"+prefix+"help cache` for more..\n").
		AddField("", "\n\n").
		SetThumbnail(avatarURL).
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName + " (build date " + version.BuildDate + ")").
		MessageEmbed

	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Fatal("Error sending embed message", err)
	}
}

func (d *Discord) handleHelpPlay() {
	prefix := d.CommandPrefix

	commandsTitle := "ℹ️ **Playback Commands**\n\n"
	command1 := fmt.Sprintf("`%vplay [title]` — play track by title (replace `[title]` with track name)\n", prefix)
	command2 := fmt.Sprintf("`%vplay [url]` — play track by URL (replace `[url]` with track URL)\n", prefix)
	command3 := fmt.Sprintf("`%vplay [stream]` — play stream by URL (replace `[stream]` with stream URL)\n", prefix)
	command4 := fmt.Sprintf("`%vplay [id]` — play track by history ID (replace `[id]` with track ID, use `%vhistory` to get ID)\n", prefix, prefix)
	command5 := fmt.Sprintf("`%vskip` — play next track\n", prefix)
	command6 := fmt.Sprintf("`%vpause` — pause playback\n", prefix)
	command7 := fmt.Sprintf("`%vresume` — resume playback\n", prefix)
	command8 := fmt.Sprintf("`%vstop` — stop playback, clear queue and leave voice channel\n", prefix)

	separator := "\n\n"

	exampleTitle := "Examples:\n"
	example1 := fmt.Sprintf("```%vplay Never Gonna Give You Up```", prefix)
	example2 := fmt.Sprintf("```%vplay https://www.youtube.com/watch?v=dQw4w9WgXcQ```", prefix)
	example3 := fmt.Sprintf("```%vplay http://stream.radioparadise.com/aac-128```", prefix)
	example4 := fmt.Sprintf("```%vplay 123``` (assuming track ID in %vhistory is 123)", prefix, prefix)

	info := "⚠️ Spotify links are not supported"

	d.sendMessageEmbed(commandsTitle + command1 + command2 + command3 + command4 + command5 + command6 + command7 + command8 + separator + exampleTitle + example1 + example2 + example3 + example4 + separator + info)
}

func (d *Discord) handleHelpQueue() {
	prefix := d.CommandPrefix

	commandsTitle := "ℹ️ **Queue Commands**\n\n"
	command1 := fmt.Sprintf("`%vadd [title]` — add track by title (replace `[title]` with track name)\n", prefix)
	command2 := fmt.Sprintf("`%vadd [url]` — add track by URL (replace `[url]` with track URL)\n", prefix)
	command3 := fmt.Sprintf("`%vadd [stream]` — add stream by URL (replace `[stream]` with stream URL)\n", prefix)
	command4 := fmt.Sprintf("`%vadd [id]` — add track by history ID (replace `[id]` with track ID, use `%vhistory` to get ID)\n", prefix, prefix)
	command5 := fmt.Sprintf("`%vlist` — show current queue\n", prefix)

	separator := "\n\n"

	exampleTitle := "Examples:\n"
	example1 := fmt.Sprintf("```%vadd Never Gonna Give You Up```", prefix)
	example2 := fmt.Sprintf("```%vadd https://www.youtube.com/watch?v=dQw4w9WgXcQ```", prefix)
	example3 := fmt.Sprintf("```%vadd http://stream.radioparadise.com/aac-128```", prefix)
	example4 := fmt.Sprintf("```%vadd 123``` (assuming track ID in %vhistory is 123)", prefix, prefix)
	example5 := fmt.Sprintf("```%vlist```", prefix)

	d.sendMessageEmbed(commandsTitle + command1 + command2 + command3 + command4 + command5 + separator + exampleTitle + example1 + example2 + example3 + example4 + example5)

}

func (d *Discord) handleHelpHistory() {

}

func (d *Discord) handleHelpInfo() {

}

func (d *Discord) handleHelpManage() {

}

func (d *Discord) handleHelpCache() {

}
