![# Header](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](/docs/README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](/docs/README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](/docs/README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](/docs/README_JP.md)

# Melodix Player

Melodix Player is a Discord music bot that does its best, even in the presence of connection errors.

## Features Overview

The bot aims to be an easy-to-use yet powerful music player. Its key objectives include:

- Playback of single/multiple tracks or playlists from YouTube, added by title or URL.
- Playback of radio streams added via URL.
- Access to the history of previously played tracks with sorting options for play counts or duration.
- Handling playback interruptions due to network failures — Melodix will attempt to resume playback.
- Exposed Rest API to perform various tasks outside of Discord commands.
- Operation across multiple Discord servers.

![Playing Example](https://github.com/keshon/melodix-player/blob/master/assets/demo.gif)

## Download Binary

Binaries (Windows only) are available on the [Release page](https://github.com/keshon/melodix-player/releases). It is recommended to build binaries from source for the latest version.

## Discord Commands

Melodix Player supports various commands with their respective aliases to control music playback. Some commands require additional parameters:

**Commands & Aliases**:
- `play` (`p`, `>`) — Parameters: YouTube video URL, history ID, track title, or valid stream link
- `skip` (`next`, `ff`, `>>`)
- `pause` (`!`)
- `resume` (`r`,`!>`)
- `stop` (`x`)
- `add` (`a`, `+`) — Parameters: YouTube video URL or history ID, track title, or valid stream link
- `list` (`queue`, `l`, `q`)
- `history` (`time`, `t`) — Parameters: `duration` or `count`
- `help` (`h`, `?`)
- `about` (`v`)
- `register`
- `unregister`

Commands should be prefixed with `!` by default. For instance, `!play`, `!>>`, and so on.

### Examples
To use the `play` command, provide a YouTube video title, URL, or a history ID as a parameter, e.g.:
`!play Never Gonna Give You Up` 
or 
`!p https://www.youtube.com/watch?v=dQw4w9WgXcQ` 
or 
`!> 5` (assuming `5` is an id that can be seen from history: `!history`)

For adding a song to the queue, use a similar approach:
`!add Never Gonna Give You Up` 
`!resume` (to start playing)

## Adding the Bot to a Discord Server

To add Melodix to your Discord server:

1. Create a bot at the [Discord Developer Portal](https://discord.com/developers/applications) and acquire the Bot's CLIENT_ID.
2. Use the following link: `discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID_HERE&scope=bot&permissions=36727824`
   - Replace `YOUR_CLIENT_ID_HERE` with your Bot's Client ID from step 1.
3. The Discord authorization page will open in your browser, allowing you to select a server.
4. Choose the server where you want to add Melodix and click "Authorize".
5. If prompted, complete the reCAPTCHA verification.
6. Grant Melodix the necessary permissions for it to function correctly.
7. Click "Authorize" to add Melodix to your server.

Once the bot is added, proceed to actual bot building.

## API Access and Routes

Melodix Player provides various routes for different functionalities:

### Guild Routes

- `GET /guild/ids`: Retrieve active guild IDs.
- `GET /guild/playing`: Obtain information about the currently playing track in each active guild.

### History Routes

- `GET /history`: Access the overall history of played tracks.
- `GET /history/:guild_id`: Fetch the history of played tracks for a specific guild.

### Avatar Routes

- `GET /avatar`: List available images in the avatar folder.
- `GET /avatar/random`: Fetch a random image from the avatar folder.

### Log Routes

- `GET /log`: Show the current log.
- `GET /log/clear`: Clear the log.
- `GET /log/download`: Download the log as a file.

## Building from Sources

This project is written in the Go language, allowing it to run on a *server* or as a *local* program.

**Local Usage**
There are several scripts provided for building Melodix Player from source:
- `bash-and-run.bat` (or `.sh` for Linux): Build the debug version and execute.
- `build-release.bat` (or `.sh` for Linux): Build the release version.
- `assemble-dist.bat`: Build the release version and assemble it as a distribution package (Windows only, UPX packager will be downloaded during the process).

For local usage, run these scripts for your operating system and rename `.env.example` to `.env`, storing your Discord Bot Token in the `DISCORD_BOT_TOKEN` variable. Install [FFMPEG](https://ffmpeg.org/) (only the recent version is supported). If your FFMPEG installation is portable, specify the path in the `DCA_FFMPEG_BINARY_PATH` variable.

**Server Usage**
To build and deploy the bot in a Docker environment, refer to the `docker/README.md` for specific instructions.

Once the binary file is built, the `.env` file is filled, and the Bot is added to your server, Melodix is ready for operation.

## Where to Get Support or Gently Pats

If you have any questions, you can ask me in my [Discord server](https://discord.gg/NVtdTka8ZT) to get support. Bear in mind there is no community whatsoever — just me.

## Acknowledgment

I drew inspiration from [Muzikas](https://github.com/FabijanZulj/Muzikas), a user-friendly Discord bot created by Fabijan Zulj.

As a result of Melodix development, a new project was born: [Discord Bot Boilerplate](https://github.com/keshon/discord-bot-boilerplate) — a framework for building Discord bots.

## License

Melodix is licensed under the [MIT License](https://opensource.org/licenses/MIT).
