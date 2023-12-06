package discord

import (
	"fmt"
	"os"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/music/utils"
)

// handleHelpCommand handles the help command for Discord.
func (d *Discord) handleHelpCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	var hostname string
	if os.Getenv("HOST") == "" {
		hostname = config.RestHostname
	} else {
		hostname = os.Getenv("HOST") // from docker environment
	}

	avatarUrl := utils.InferProtocolByPort(hostname, 443) + hostname + "/avatar/random?" + fmt.Sprint(time.Now().UnixNano())
	slog.Info(avatarUrl)

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
		AddField("", "*Adinistration*\n"+register+unregister).
		SetThumbnail(avatarUrl). // TODO: move out to config .env file
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}
