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

	help := fmt.Sprintf("`%vhelp`, `%vh` — show help\n", prefix, prefix)
	about := fmt.Sprintf("`%vabout`, `%vv` — show version\n", prefix, prefix)
	now := fmt.Sprintf("`%vnow` — show currently playing track name\n", prefix)

	cached := fmt.Sprintf("`%vcached` — show cached tracks\n", prefix)
	cachedSync := fmt.Sprintf("`%vcached sync` — sync added/removed files to with database\n", prefix)
	curl := fmt.Sprintf("`%vcurl [url]` — cache track (youtube url only)\n", prefix)
	uploaded := fmt.Sprintf("`%vuploaded` — show uploaded videos\n", prefix)
	uploadedExtract := fmt.Sprintf("`%vuploaded extract` — extract audio from uploaded videos to cache\n", prefix)

	register := fmt.Sprintf("`%vregister` — enable commands listening\n", prefix)
	unregister := fmt.Sprintf("`%vunregister` — disable commands listening\n", prefix)
	whoami := fmt.Sprintf("`%vwhoami` — log user's info\n", prefix)
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
		AddField("", "**History**\n"+history+historyByDuration+historyByPlaycount+"\n").
		AddField("", "").
		AddField("", "**Information**\n"+now+help+about+"\n").
		AddField("", "").
		AddField("", "**Management**\n"+register+unregister+whoami+melodixPrefix+melodixPrefixUpdate+melodixPreifxReset+"\n").
		AddField("", "").
		AddField("", "**Caching & Sideloading**\nThis commands are for superadmin only.\n"+cached+cachedSync+curl+uploaded+uploadedExtract+"\n").
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

	command1 := fmt.Sprintf("▶️ **Play**\n`%vplay [title|url|stream|id]`\n", prefix)
	command2 := fmt.Sprintf("`%vp [title|url|stream|id]`\n", prefix)
	command3 := fmt.Sprintf("`%v> [title|url|stream|id]`\n\n", prefix)
	command5 := fmt.Sprintf("⏭️ **Skip**\n`%vskip`\n`%vnext`\n`%v>>`\n\n", prefix, prefix, prefix)
	command6 := fmt.Sprintf("⏸ **Pause**\n`%vpause`\n`%v!`\n\n", prefix, prefix)
	command7 := fmt.Sprintf("⏯️	**Resume**\n`%vresume`\n`%vr`\n`%v!>`\n\n", prefix, prefix, prefix)
	command8 := fmt.Sprintf("⏹️ **Stop**\n`%vstop`\n`%vx`\n\n", prefix, prefix)

	exampleTitle := "▬ Examples ▬▬▬▬▬▬\n"

	example1 := fmt.Sprintf("```%vplay Never Gonna Give You Up```", prefix)
	example2 := fmt.Sprintf("```%vp https://www.youtube.com/watch?v=dQw4w9WgXcQ```", prefix)
	example3 := fmt.Sprintf("```%vp https://www.youtube.com/watch?v=dQw4w9WgXcQ https://www.youtube.com/watch?v=98MWcF_Ucs0``` (multiple links added, space separated)", prefix)
	example4 := fmt.Sprintf("```%v> http://stream.radioparadise.com/aac-128```", prefix)
	example5 := fmt.Sprintf("```%vplay 123``` (assuming track ID in %vhistory is 123)", prefix, prefix)

	info1 := "title - is a song title, url - YouTube URL, stream - valid stream URL (radio), id - track id from *History*\n\n"
	info2 := "\n\n⚠️ Spotify links are not supported"

	d.sendMessageEmbed(command1 + command2 + command3 + info1 + command5 + command6 + command7 + command8 + exampleTitle + example1 + example2 + example3 + example4 + example5 + info2)
}

func (d *Discord) handleHelpQueue() {
	prefix := d.CommandPrefix

	command1 := fmt.Sprintf("🆕 **Add to queue**\n`%vadd [title|url|stream|id]`\n", prefix)
	command2 := fmt.Sprintf("`%va [title|url|stream|id]`\n", prefix)
	command3 := fmt.Sprintf("`%v+ [title|url|stream|id]`\n\n", prefix)
	command4 := fmt.Sprintf("📑 **Show queue**\n`%vlist`\n`%vl`\n`%vq`\n\n", prefix, prefix, prefix)

	exampleTitle := "▬ Examples ▬▬▬▬▬▬\n"

	example1 := fmt.Sprintf("```%vadd Never Gonna Give You Up```", prefix)
	example2 := fmt.Sprintf("```%va Never Gonna Give You Up```", prefix)
	example3 := fmt.Sprintf("```%v+ https://www.youtube.com/watch?v=dQw4w9WgXcQ```", prefix)
	example4 := fmt.Sprintf("```%vlist```", prefix)
	example5 := fmt.Sprintf("```%vl```", prefix)
	example6 := fmt.Sprintf("```%vq```", prefix)

	d.sendMessageEmbed(command1 + command2 + command3 + command4 + exampleTitle + example1 + example2 + example3 + example4 + example5 + example6)

}

func (d *Discord) handleHelpHistory() {

}

func (d *Discord) handleHelpInfo() {

}

func (d *Discord) handleHelpManage() {

}

func (d *Discord) handleHelpCache() {

}
