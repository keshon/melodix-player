package melodix

import (
	"encoding/base64"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

// parseCommand parses the command and parameter from the Discord input based on the provided pattern.
func parseCommand(content, pattern string) (string, string, error) {
	if !strings.HasPrefix(content, pattern) {
		return "", "", fmt.Errorf("pattern not found")
	}

	content = content[len(pattern):] // Strip the pattern

	words := strings.Fields(content) // Split by whitespace, handling multiple spaces
	if len(words) == 0 {
		return "", "", fmt.Errorf("no command found")
	}

	command := strings.ToLower(words[0])
	parameter := ""
	if len(words) > 1 {
		parameter = strings.Join(words[1:], " ")
		parameter = strings.TrimSpace(parameter)
	}
	return command, parameter, nil
}

// getCanonicalCommand gets the canonical command from aliases using the given alias.
func getCanonicalCommand(alias string, commandAliases [][]string) string {
	for _, aliases := range commandAliases {
		for _, a := range aliases {
			if a == alias {
				return aliases[0]
			}
		}
	}
	return ""
}

// parseSongsAndTypeInParameter parses the type and parameters from the input parameter string.
func parseSongsAndTypeInParameter(param string) (string, []string) {
	// Trim spaces at the beginning and end
	param = strings.TrimSpace(param)

	if len(param) == 0 {
		return "", []string{}
	}

	// Check if the parameter is a URL
	u, err := url.Parse(param)
	if err == nil && (u.Scheme == "http" || u.Scheme == "https") && isYouTubeURL(u.Host) {
		// If it's a URL, split by ",", " ", new line, or carriage return
		paramSlice := strings.FieldsFunc(param, func(r rune) bool {
			return r == ',' || r == '\n' || r == '\r' || r == ' ' || r == '\t'
		})
		return "url", paramSlice
	}

	// Check if the parameter is an ID
	params := strings.Fields(param)
	allValidIDs := true
	for _, param := range params {
		_, err := strconv.Atoi(param)
		if err != nil {
			allValidIDs = false
			break
		}
	}
	if allValidIDs {
		return "id", params
	}

	// Treat it as a single title if it's not a URL or ID
	return "title", []string{param}
}

// isYouTubeURL checks if the host is a YouTube URL.
func isYouTubeURL(host string) bool {
	return host == "www.youtube.com" || host == "youtube.com" || host == "youtu.be"
}

// parseVideoParamsFromYoutubeURL parses video parameters from a YouTube URL.
func parseVideoParamsFromYouTubeURL(urlString string) (duration float64, contentLength int, expiryTimestamp int64, err error) {
	duration = -1
	contentLength = -1
	expiryTimestamp = -1

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		return duration, contentLength, expiryTimestamp, fmt.Errorf("failed to parse URL: %v", err)
	}

	queryParams := parsedURL.Query()

	durParam, err := strconv.ParseFloat(queryParams.Get("dur"), 64)
	if err != nil {
		duration = -1
	}
	duration = durParam

	if clenParam := queryParams.Get("clen"); clenParam != "" {
		contentLength, err = strconv.Atoi(clenParam)
		if err != nil {
			return duration, contentLength, expiryTimestamp, fmt.Errorf("failed to parse content length: %v", err)
		}
	}

	if expireParam := queryParams.Get("expire"); expireParam != "" {
		expiryTimestamp, err = strconv.ParseInt(expireParam, 10, 64)
		if err != nil {
			return duration, contentLength, expiryTimestamp, fmt.Errorf("failed to parse expiry timestamp: %v", err)
		}
	}

	return duration, contentLength, expiryTimestamp, nil
}

// String returns the string representation of the CurrentStatus.
func (status Status) String() string {
	statuses := map[Status]string{
		StatusResting: "Resting",
		StatusPlaying: "Playing",
		StatusPaused:  "Paused",
		StatusError:   "Error",
	}

	return statuses[status]
}

// formatDuration formats the given seconds into HH:MM:SS format.
func formatDuration(seconds float64) string {
	totalSeconds := int(seconds)
	hours := totalSeconds / 3600
	totalSeconds %= 3600
	minutes := totalSeconds / 60
	seconds = math.Mod(float64(totalSeconds), 60)
	return fmt.Sprintf("%02d:%02d:%02.0f", hours, minutes, seconds)
}

// getRandomAvatarPath returns path to randomly selected file in specified folder
func getRandomImagePathFromPath(folderPath string) (string, error) {

	var validFiles []string
	files, err := os.ReadDir(folderPath)
	if err != nil {
		return "", err
	}

	// Filter only files with certain extensions (you can modify this if needed)
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jpg" || filepath.Ext(file.Name()) == ".png" {
			validFiles = append(validFiles, file.Name())
		}
	}

	if len(validFiles) == 0 {
		return "", fmt.Errorf("no valid images found")
	}

	// Get a random index
	randomIndex := rand.Intn(len(validFiles))
	randomImage := validFiles[randomIndex]
	imagePath := filepath.Join(folderPath, randomImage)

	return imagePath, nil
}

func readFileToBase64(imgPath string) (string, error) {
	img, err := os.ReadFile(imgPath)
	if err != nil {
		return "", fmt.Errorf("error reading the response: %v", err)
	}

	base64Img := base64.StdEncoding.EncodeToString(img)
	return fmt.Sprintf("data:%s;base64,%s", http.DetectContentType(img), base64Img), nil
}

func sanitizeString(input string) string {
	// Define a regular expression to match unwanted characters
	re := regexp.MustCompile("[[:^print:]]")

	// Replace unwanted characters with an empty string
	sanitized := re.ReplaceAllString(input, "")

	return sanitized
}

func getRandomWaitPhrase() string {
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

func getRandomAboutTitlePhrase() string {
	phrases := []string{
		"Hello there!",
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

func getRandomAboutDescriptionPhrase() string {
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

// inferProtocolByPort attempts to infer the protocol based on the availability of a specific port.
func inferProtocolByPort(hostname string, port int) string {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", hostname, port))
	if err != nil {
		// Assuming it's not available, use HTTP
		return "http://"
	}
	defer conn.Close()

	// The port is available, use HTTPS
	return "https://"
}
