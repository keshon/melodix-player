![# Header](https://github.com/keshon/melodix-discord-player/blob/master/assets/banner.jpg)


# Melodix Discord Player

Melodix is a Discord music bot that just works even if connection errors are along the way.


## Features Overview

The bot aims to be an easy-to-use yet powerful music player beastie. Its key objectives include:

- Playback single/multiple tracks or playlists from Youtube added by title or URL.
- Playback of radio streams added via URL.
- Access the history of previously played tracks with sorting by play counts or duration.
- Handling playback interruptions due to network failures — Melodix will try to resume the playback.
- Exposed Rest API to do various magic tasks outside of Discord commands.
- Operate across multiple Discord servers.

![# Playing Example](https://github.com/keshon/melodix-discord-player/blob/master/assets/playing.jpg)


## Download binary

Binaries (Windows only) are available at [Release page](https://github.com/keshon/melodix-discord-player/releases).
It is recommended to build binaries from source if you want to have the latest version.


## Discord Commands

Melodix supports various commands with their respective aliases to control music playback. Some commands require additional parameters:

- Commands & Aliases:
  — `play` (`p`, `>`) — Parameters: YouTube video URL, history ID, or track title
  — `skip` (`next`, `ff`, `>>`)
  — `pause` (`!`)
  — `resume` (`r`,`!>`)
  — `stop` (`x`)
  — `add` (`a`, `+`) — Parameters: YouTube video URL or history ID, or track title
  — `list` (`queue`, `l`, `q`)
  — `history` (`time`, `t`) — Parameters: `duration` or `count`
  — `help` (`h`, `?`)
  — `about` (`v`)
  — `register`
  — `unregister`

Commands should be prefixed with `!` by default. For instance, `!play`, `!>>`, and so on.

To use the `play` and `add` commands, provide a YouTube video title, URL, or a history ID as a parameter, e.g.:
`!play Never Gonna Give You Up` 
or 
`!p https://www.youtube.com/watch?v=dQw4w9WgXcQ` 
or 
`!> 5` (assuming `5` is an id that can be seen from history: `!history`)

Similarly, for adding a song to the queue, use a similar approach.


## Adding the Bot to a Discord Server

To add Melodix to your Discord server:

1. Create a bot at the [Discord Developer Portal](https://discord.com/developers/applications) and acquire the Bot's CLIENT_ID.
2. Use the following link: `discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID_HERE&scope=bot&permissions=36727824`
   — Replace `YOUR_CLIENT_ID_HERE` with your Bot's Client ID from step 1.
3. The Discord authorization page will open in your browser, allowing you to select a server.
4. Choose the server where you want to add Melodix and click "Authorize".
5. If prompted, complete the reCAPTCHA verification.
6. Grant Melodix the necessary permissions for it to function correctly.
7. Click "Authorize" to add Melodix to your server.

Once 'bot' is added proceed to actual bot building.

## API Access and Routes

Melodix provides various routes for different functionalities:

#### Guild Routes

- `GET /guild/ids`: Retrieve active guild IDs.
- `GET /guild/playing`: Obtain information about the currently playing track in each active guild.

#### History Routes

- `GET /history`: Access the overall history of played tracks.
- `GET /history/:guild_id`: Fetch the history of played tracks for a specific guild.

#### Avatar Routes

- `GET /avatar`: List available images in avatar folder.
- `GET /avatar/random`: Fetch random image from avatar folder.

#### Log Routes

- `GET /log`: Show current log.
- `GET /log/clear`: Clear log.
- `GET /log/download`: Download log as a file.

## Building from sources

Melodix is written in Go language, allowing it to run on a *server* or as a *local* program.

**Local Usage**
There are several scripts provided for building Melodix:
  — `bash-and-run.bat` (or `.sh` for Linux): Build the debug version and execute.
  — `build-release.bat` (or `.sh` for Linux): Build the release version. Note: The UPX packer is called as a final step; if not installed, comment it out.

For local usage, run these scripts for your operating system and rename `.env.example` to `.env`, storing your Discord Bot Token in the `DISCORD_BOT_TOKEN` variable.
Install [FFMPEG](https://ffmpeg.org/) (only recent version is supported). If your FFMPEG installation is portable specify path in the `DCA_FFMPEG_BINARY_PATH` variable.

**Server Usage**
To build and deploy the bot in a Docker environment refer to the `deploy/README.md` for specific instructions.

Once the binary file is built, the `.env` file is filled, and the Bot is added to your server, Melodix is ready for operation.

## Where to get gently pats

If you have any questions you can ask me in my [Discord server](https://discord.gg/NVtdTka8ZT) to get support. Bear in mind there is no community whatsoever — just me.

## Acknowledgment

I drew inspiration from [Muzikas](https://github.com/FabijanZulj/Muzikas), a user-friendly Discord bot created by Fabijan Zulj.

As a result of Melodxi development, a new project was born: [Discord Bot Boilerplate](https://github.com/keshon/discord-bot-boilerplate) — a framework for building Discord bots .

## License

Melodix is licensed under the [MIT License](https://opensource.org/licenses/MIT).
