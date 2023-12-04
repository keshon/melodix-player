package discord

import (
	"fmt"
	"math/rand"
	"time"

	embed "github.com/Clinet/discordgo-embed"
	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"

	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/internal/version"
	"github.com/keshon/melodix-discord-player/music/utils"
)

// handleAboutCommand handles the about command for Discord.
func (d *Discord) handleAboutCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.changeAvatar(s)

	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	avatarUrl := utils.InferProtocolByPort(config.RestHostname, 443) + config.RestHostname + "/avatar/random?" + fmt.Sprint(time.Now().UnixNano())
	slog.Info(avatarUrl)

	title := GetRandomAboutTitlePhrase()
	content := GetRandomAboutDescriptionPhrase()

	embedStr := fmt.Sprintf("**%v**\n\n%v", title, content)

	embedMsg := embed.NewEmbed().
		SetDescription(embedStr).
		AddField("```"+version.BuildDate+"```", "Build date").
		AddField("```"+version.GoVersion+"```", "Go version").
		AddField("```Created by Innokentiy Sokolov```", "[Linkedin](https://www.linkedin.com/in/keshon), [GitHub](https://github.com/keshon), [Homepage](https://keshon.ru)").
		InlineAllFields().
		SetImage(avatarUrl).
		SetColor(0x9f00d4).SetFooter(version.AppFullName).MessageEmbed

	s.ChannelMessageSendEmbed(m.Message.ChannelID, embedMsg)
}

func GetRandomAboutTitlePhrase() string {
	phrases := []string{
		"Well, hello there!",
		"Who do we have here?",
		"Brace yourselves for Melodix!",
		"Get ready to laugh and groove!",
		"Peek behind the musical curtain!",
		"Unleashing Melodix magic!",
		"Prepare for some bot banter!",
		"It's showtime with Melodix!",
		"Allow me to introduce myself",
		"Heeeey amigos!",
		"Unleashing Melodix magic!",
		"Did someone order beats?",
		"Well, look who's curious!",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}

func GetRandomAboutDescriptionPhrase() string {
	phrases := []string{
		"🎶 The Discord DJ That Won't Take Requests From Your In-Laws! 🔊 Crank up the tunes and drown out the chaos. No commercials, no cover charges—just pure, unfiltered beats. Because when life hands you a mic, you drop it with Melodix! 🎤🎉 #MelodixMadness #NoRequestsAllowed",
		"🎵 Groovy Bot: Where Beats Meet Banter! 🤖 Tune in for the ultimate audio fiesta. Tracks that hit harder than Monday mornings and a vibe that won't quit. Request, rewind, and revel in the groove. Life's a party; let's make it legendary! 🚀🕺 #GroovyBot #UnleashTheBeats",
		"Melodix: Unleash the Epic Beats! 🚀🎵 Your Discord, Your Soundtrack—Elevate your server experience with the ultimate music companion. No boundaries, just epicness! Turn up the volume and let Melodix redefine your sonic adventure. 🎧🔥 #EpicBeats #MelodixUnleashed",
		"🤖 Welcome to the Groovy Bot Experience! 🎶 Unleash the musical mayhem with a sprinkle of humor. I'm your DJ, serving beats hotter than a summer grill. 🔥 Request a jam, peek into your play history, and let's dance like nobody's watching. It's music with a side of laughter – because why not? Let the groove take the wheel! 🕺🎉 #BotLife #DanceTillYouDrop",
		"🎶 Melodix: Your Personal Discord DJ! 🔊 I spin tunes better than your grandma spins knitting yarn. No song requests? No problem! I play what I want, when I want. Get ready for a musical rollercoaster, minus the safety harness! 🎢🎤 #MelodixMagic #GrandmaApproved",
		"🎵 Melodix: The Bot with the Moves! 🕺 Break out your best dance moves because I'm dropping beats that even the neighbors can't resist. Turn up the volume, lock the door, and dance like nobody's watching—except me, of course! 💃🎉 #DanceFloorOnDiscord #BeatDropper",
		"Melodix: Where Music Meets Mischief! 🤖🎶 Your server's audio adventure begins here. I play music that hits harder than your morning alarm and cracks more jokes than your favorite stand-up comedian. Buckle up; it's gonna be a hilarious ride! 🚀😂 #MusicMischief #JokesterBot",
		"🤖 Meet Melodix: The Discord DJ on a Comedy Tour! 🎤 Unleash the laughter and the beats with a bot that's funnier than your uncle's dad jokes. Request a track, sit back, and enjoy the show. Warning: I may cause uncontrollable fits of joy! 😆🎵 #ComedyTourBot #LaughOutLoud",
		"🎧 Melodix: Beats that Hit Harder Than Life's Problems! 💥 When reality knocks, I turn up the volume. Melodix delivers beats that punch harder than Monday mornings and leave you wondering why life isn't always this epic. Buckle up; it's time to conquer the airwaves! 🚀🎶 #EpicBeats #LifePuncher",
		"🔊 Groovy Bot: Making Discord Groovy Again! 🕺 Shake off the stress, kick back, and let Groovy Bot do the heavy lifting. My beats are so groovy; even your grandma would break into the moonwalk. Get ready to rediscover your groove on Discord! 🌙💫 #GroovyAgain #DiscordDanceRevolution",
		"🚀 Melodix: Your Gateway to Musical Awesomeness! 🌟 I'm not just a bot; I'm your VIP pass to a sonic wonderland. No queues, no limits—just pure, unadulterated musical awesomeness. Fasten your seatbelts; the journey to epic sounds begins now! 🎸🎉 #MusicalAwesomeness #VIPPass",
		"🎶 Melodix: More Than Just a Bot—It's a Vibe! 🤖🕶️ Elevate your server with vibes so cool, even penguins envy me. I'm not your average bot; I'm a mood-altering, vibe-creating, beat-dropping phenomenon. Prepare for a vibe check, Melodix style! 🌊🎵 #VibeMaster #BotGoals",
		"🔊 Step into Melodix's Audio Playground! 🎉 Your ticket to the ultimate sonic adventure is here. With beats that rival a theme park ride and humor sharper than a stand-up special, Melodix is your all-access pass to the audio amusement park. Let the fun begin! 🎢🎤 #AudioPlayground #RollercoasterBeats",
		"🎵 Melodix: Where Discord Gets Its Groove On! 💃 I'm not just a bot; I'm the rhythm that keeps your server dancing. My beats are so infectious; even the toughest critics tap their feet. Get ready to groove; Melodix is in the house! 🕺🎶 #DiscordGrooveMaster #BeatCommander",
		"🚀 Unleash Melodix: The Bot with a Sonic Punch! 💥 Dive into a world where beats hit harder than a superhero landing. Melodix isn't just a bot; I'm a powerhouse of sonic awesomeness. Get ready for an audio experience that packs a punch! 🎤👊 #SonicPowerhouse #BeatHero",
		"🔊 Melodix: Your Server's Audio Magician! 🎩✨ Watch as I turn ordinary moments into extraordinary memories with a wave of my musical wand. Beats appear, laughter ensues, and your server becomes the stage for an epic audio performance. Prepare to be enchanted! 🎶🔮 #AudioMagician #DiscordWizard",
		"🎧 Melodix: Beats That Speak Louder Than Words! 📢 When words fail, music speaks. I deliver beats so powerful; even a whisper could start a party. Say goodbye to silence; it's time to let the music do the talking. Turn it up; let's break the sound barrier! 🚀🎵 #BeatsNotWords #MusicSpeaksVolumes",
		"🤖 Melodix: The Bot That Takes the Stage! 🎤 Roll out the red carpet; Melodix is here to steal the show. My beats command attention, and my humor steals the spotlight. It's not just music; it's a performance. Get ready for a standing ovation! 👏🎶 #StageStealer #BotOnTheMic",
		"🎵 Groovy Bot: Turning Discord into a Dance Floor! 💃 I'm not just a bot; I'm the DJ that turns your server into a non-stop dance party. Groovy Bot's beats are so infectious; even the furniture wants to boogie. Get ready to dance like nobody's watching! 🎉🎶 #DancePartyBot #BoogieMaster",
		"🚀 Melodix: Your Sonic Co-Pilot on the Discord Journey! 🎶 Buckle up; we're about to take off on a musical adventure. Melodix isn't just a bot; I'm your co-pilot navigating the airspace of epic beats. Fasten your seatbelts; the journey awaits! ✈️🔊 #SonicCoPilot #DiscordAdventure",
		"🔊 Melodix: Bringing the Beats, Igniting the Vibes! 🔥 I'm not just a bot; I'm the ignition switch for a server-wide party. My beats are so fire; even the speakers need a cooldown. Prepare for a musical blaze that'll leave you in awe! 🎵🎉 #IgniteTheVibes #DiscordInferno",
		"🎶 Melodix: Turning Mundane into Musical! 🌟 Say goodbye to the ordinary; Melodix is here to transform the mundane into a symphony of epic proportions. My beats are the soundtrack to your server's extraordinary journey. Let's make every moment musical! 🎤🚀 #MusicalTransformation #EpicSymphony",
		"🤖 Melodix: The Bot That Doesn't Miss a Beat—Literally! 🥁 Precision beats, flawless execution, and humor that lands every time. Melodix is the maestro of your server's audio orchestra. No missed beats, no dull moments—just pure musical perfection! 🎶👌 #NoMissedBeats #AudioMaestro",
		"🎵 Groovy Bot: Where Discord Finds Its Rhythm! 🕺 We're not just a bot; we're the rhythm that keeps your server in sync. Groovy Bot's beats are so contagious; even the skeptics catch the vibe. Get ready for a rhythmic revolution on Discord! 🎶🔄 #RhythmicRevolution #DiscordSyncMaster",
		"🚀 Melodix: Elevate Your Discord, Elevate Your Beats! 🎧 We're not just a bot; we're the elevator to the next level of sonic greatness. Melodix's beats are the soundtrack to your server's ascension. Get ready to elevate your vibes to new heights! 🌌🔊 #ElevateYourBeats #DiscordAscent",
	}

	index := rand.Intn(len(phrases))

	return phrases[index]
}
