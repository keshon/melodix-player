package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/gookit/slog"
	"github.com/keshon/melodix-discord-player/internal/config"
	"github.com/keshon/melodix-discord-player/music/pkg/dca"
)

// handleHelpCommand handles the help command for Discord.
func (d *Discord) handleRadioCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	channel, err := s.State.Channel(m.Message.ChannelID)
	if err != nil {
		slog.Error(err)
	}

	guild, err := s.State.Guild(channel.GuildID)
	if err != nil {
		slog.Error(err)
	}

	vs, found := findUserVoiceState(m.Message.Author.ID, guild.VoiceStates)
	if !found {
		slog.Error("user not found in voice channel")
	}

	conn, err := d.Session.ChannelVoiceJoin(channel.GuildID, vs.ChannelID, false, true)

	config, err := config.NewConfig()
	if err != nil {
		slog.Fatalf("Error loading config: %v", err)
	}

	options := &dca.EncodeOptions{
		Volume:                  1.0,
		FrameDuration:           config.DcaFrameDuration,
		Bitrate:                 config.DcaBitrate,
		PacketLoss:              config.DcaPacketLoss,
		RawOutput:               config.DcaRawOutput,
		Application:             config.DcaApplication,
		CompressionLevel:        config.DcaCompressionLevel,
		BufferedFrames:          config.DcaBufferedFrames,
		VBR:                     config.DcaVBR,
		StartTime:               0,
		ReconnectAtEOF:          config.DcaReconnectAtEOF,
		ReconnectStreamed:       config.DcaReconnectStreamed,
		ReconnectOnNetworkError: config.DcaReconnectOnNetworkError,
		ReconnectOnHttpError:    config.DcaReconnectOnHttpError,
		ReconnectDelayMax:       config.DcaReconnectDelayMax,
		FfmpegBinaryPath:        config.DcaFfmpegBinaryPath,
		EncodingLineLog:         config.DcaEncodingLineLog,
		UserAgent:               config.DcaUserAgent,
	}

	// Open the radio station stream
	options.RawOutput = true
	options.Bitrate = 96
	options.Application = "lowdelay"

	radioStationURL := "http://ipx.psyradio.org:8010/"
	encodeSession, err := dca.EncodeFile(radioStationURL, options)
	if err != nil {
		fmt.Println("Error encoding radio stream:", err)
		return
	}

	defer encodeSession.Cleanup()

	// Send the audio to Discord
	done := make(chan error)
	dca.NewStream(encodeSession, conn, done)

	// Wait for the stream to finish or an error to occur
	err = <-done
	if err != nil && err != dca.ErrVoiceConnClosed {
		fmt.Println("Error streaming audio:", err)
	}

	// Disconnect from the voice channel
	conn.Disconnect()
}
