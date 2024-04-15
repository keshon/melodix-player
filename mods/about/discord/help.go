package discord

import (
	"fmt"
	"os"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-player/internal/config"
	"github.com/keshon/melodix-player/internal/version"
	"github.com/keshon/melodix-player/mods/about/utils"
)

// handleHelpCommand handles the help command for the Discord bot.
//
// Takes in a session and a message create, and does not return any value.
func (d *Discord) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
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
	prefix := d.CommandPrefix

	play := fmt.Sprintf("`%vplay [title/url/id/stream]` — play selected track/radio\n", prefix)
	skip := fmt.Sprintf("`%vskip` — play next track\n", prefix)
	pause := fmt.Sprintf("`%vpause`, `%vresume` — pause/resume playback\n", prefix, prefix)
	stop := fmt.Sprintf("`%vstop` — stop playback and leave voice channel\n", prefix)

	queue := fmt.Sprintf("`%vadd [title/url/id]` — add track\n", prefix)
	list := fmt.Sprintf("`%vlist` — show current queue\n", prefix)

	history := fmt.Sprintf("`%vhistory` — show played tracks\n", prefix)
	historyByDuration := fmt.Sprintf("`%vhistory duration` — sort by duration \n", prefix)
	historyByPlaycount := fmt.Sprintf("`%vhistory count` — sort by play count \n\n", prefix)

	cached := fmt.Sprintf("`%vcached` — show chached tracks\n", prefix)
	curl := fmt.Sprintf("`%vcurl [url]` — cache track (youtube only)\n", prefix)
	uploaded := fmt.Sprintf("`%vuploaded` — show uploaded videos\n", prefix)
	uploadedExtract := fmt.Sprintf("`%vuploaded extract` — cache audio from uploaded videos\n", prefix)

	help := fmt.Sprintf("`%vhelp`, `%vh` — show help\n", prefix, prefix)
	about := fmt.Sprintf("`%vabout`, `%vv` — show version\n", prefix, prefix)
	register := fmt.Sprintf("`%vregister` — enable commands listening\n", prefix)
	unregister := fmt.Sprintf("`%vunregister` — disable commands listening", prefix)

	embedMsg := embed.NewEmbed().
		SetTitle(fmt.Sprintf("ℹ️ %v — Command Usage", version.AppName)).
		SetDescription("Some commands are aliased for shortness.\n\n*[title]* - track name\n*[url]* - YouTube URL\n*[id]* - track id from *History*\n*[stream]* - valid stream URL (radio).").
		AddField("", "**Playback**\n"+play+skip+pause+stop).
		AddField("", "").
		AddField("", "**Queue**\n"+queue+list).
		AddField("", "").
		AddField("", "**History**\n"+history+historyByDuration+historyByPlaycount).
		AddField("", "").
		AddField("", "**Caching**\n"+cached+curl+uploaded+uploadedExtract).
		AddField("", "").
		AddField("", "**Information**\n"+help+about).
		AddField("", "").
		AddField("", "**Managing Bot**\n"+register+unregister).
		SetThumbnail(avatarURL).
		SetColor(0x9f00d4).
		SetFooter(version.AppFullName).
		MessageEmbed

	_, err = s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
	if err != nil {
		slog.Fatal("Error sending embed message", err)
	}
}
