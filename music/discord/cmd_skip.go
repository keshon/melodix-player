package discord

import (
	"math/rand"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
)

// handleSkipCommand handles the skip command for Discord.
func (d *Discord) handleSkipCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	embedStr := "⏩ " + getSkipPhrase()
	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		SetColor(0x9f00d4).MessageEmbed
	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)

	d.Player.Skip()
}

// getSkipPhrase returns a random skip phrase.
func getSkipPhrase() string {
	phrases := []string{
		"Skipping ahead, just like a pro.",
		"Onward and upward! Skipping...",
		"Track's taking a backseat. Skip time!",
		"Fasten your seatbelts, we're skipping!",
		"Skipping ahead for your listening pleasure.",
		"Moving on to the next groove. Skip it!",
		"Adiós, old track! We're skipping away.",
		"Skipper's at the helm. Moving on!",
		"Hit the skip button! Next track, please!",
		"Skipping like a stone on a musical river.",
		"Track's getting skipped. Brace yourself!",
		"Leaping to the next musical chapter. Skip!",
		"Skipping beats, not heartbeats. Let's go!",
		"Pressing skip, because we can. Bye, track!",
		"Track's on a timeout. Skipping in 3...2...1...",
		"Jumping tracks like it's a musical hopscotch.",
		"Track's taking a detour. We're skipping lanes.",
		"Skipper, ahoy! Full speed ahead to the next track!",
		"Track's got a date with the skip button. Adieu!",
		"Skipping beats faster than a caffeinated drummer.",
		"Next track, please! We're on a skipping spree.",
		"Track's doing the shuffle. We're pressing skip!",
		"Skip-a-dee-doo-dah, skipping all the way!",
		"Track's caught the skipping fever. Join the dance!",
		"Skipping like it's a musical game of tag.",
		"Skipocalypse now! Pressing skip for the future.",
		"Track's saying farewell. We're saying skip!",
		"Skip-tastic voyage to the next musical wonderland!",
		"Track's taking a shortcut. We're pressing skip!",
		"Groovy beats ahead, skipping to the next vibe!",
		"Skippin' and flippin', Dave Chappelle style!",
		"Skipping tracks, dodging awkward convos.",
		"Track's out, like a friend who owes you.",
		"Skipping beats, not paying for streaming.",
		"Track's leaving. Heard my routine, no doubt!",
		"Track's gone, like my Monday morning motivation.",
		"Skipper's in charge. 'Nope!' Next track, please!",
		"Track's ducking out, heard my dad jokes coming.",
		"Skipping ahead, like a politician dodging questions.",
		"Next track, please! Skipping faster than memes go viral.",
		"Track's ghosting, like my ex when rent's due.",
		"Skipping beats like I skip leg day. No regrets!",
		"Track's taking a break, needs therapy. Good luck!",
		"Skipper's on deck. Ship sailing without that track.",
		"Track's leaving the party, heard my dance moves suck!",
		"Skipping ahead, like a cat avoiding water. Classic.",
		"Skipping tracks, like I skip family gatherings.",
		"Track's outta here, like people during my rants!",
		"Skipping beats, not apologies. Sorry, not sorry!",
		"Track's gone, like common sense in YouTube comments.",
		"Skipping like avoiding traffic on a Monday morning.",
		"Skipper's steering. This track ain't invited!",
		"Skipper's calling the shots. This track is benched!",
		"Skipping ahead, like someone dodging my sarcasm.",
		"Skip like a pwofeshionaw UwU gamer g-wirl~",
		"Track's AFK, time to skip and OwO!",
		"OMG, let's skip that twack, UwU!",
		"Skipper mode: Actiwated, nya~",
		"Press F to skip the twack, UwU!",
		"Track's a noobie, let's skip, OwO!",
		"Skip faster than a wag spike, hehe~",
		"Track.exe not found, so wet's skip!",
		"Gamer g-wirl says: Skip it, nya~",
		"Wevew up, time to skip the twack, UwU!",
		"Skip wike it's respawn time, hehe~",
		"Track wage quit, wet's skip, OwO!",
		"Epic skip moment, kawaii desu~",
		"Skip, the e-gamer way, UwU!",
		"Track's camping, wet's skip it, nya~",
		"Skip, GG, EZ, kawaii victory~",
		"Skip > Loot, UwU priorities!",
		"Track's no match, let's skip, nya~",
		"Skip, headshot style, so kawaii~",
		"Track's buffering, skip to victory, UwU!",
		"Skip wike you just got a wegendawy dwop, hehe~",
		"E-giww powers, actiwate! Skip, nya~",
		"Track's in a noobie wobby, let's skip it, OwO!",
		"Skip wike you're in a speedwun, UwU!",
		"Track's wast on the weadeboawd, wet's skip!",
		"Skip, because wespawn is for noobies, nya~",
		"Track's stuck in the tutowiaw, time to skip!",
		"Skip for the memes, UwU girl~",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}
