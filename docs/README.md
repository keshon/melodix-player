# ⚠️ Project Deprecated ⚠️

**Notice:** This project is no longer actively supported. Development has moved to a new repository: [Melodix](https://github.com/keshon/melodix). Please visit the new project for the latest updates and support.

![# Header](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](./README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](./README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](./README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](./README_JP.md)

# 🎵 Melodix Player — Self-hosted Discord music bot written in Go

Melodix Player is my pet project that plays audio from YouTube and audio streaming links to Discord voice channels.

![Playing Example](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/demo.gif)

## 🌟 Features Overview

### 🎧 Playback Support
- 🎶 Single track added by song name or YouTube link.
- 🎶 Multiple tracks added via multiple YouTube links (space separated).
- 🎶 Tracks from public user playlists.
- 🎶 Tracks from "MIX" playlists.
- 📻 Streaming links (e.g., radio stations).

### ⚙️ Additional Features
- 🌐 Operation across multiple Discord servers (guild management).
- 📜 Access to history of previously played tracks with sorting options.
- 💾 Downloading tracks from YouTube as mp3 files for caching.
- 🎼 Sideloading audio mp3 files.
- 🎬 Sideloading video files with audio extraction as mp3 files.
- 🔄 Playback auto-resume support for connection interruptions.
- 🛠️ REST API support (limited at the moment).

### ⚠️ Current Limitations
- 🚫 The bot cannot play YouTube streams.
- ⏸️ Playback auto-resume support creates noticeable pauses.
- ⏩ Sometimes playback speed is slightly faster than intended.
- 🐞 It's not bug-free.

## 🚀 Try Melodix Player

You can test Melodix in two ways:
- 🖥️ Download [compiled binaries](https://github.com/keshon/melodix-player/releases) (available only for Windows). Ensure FFMPEG is installed on your system and added to the global PATH variable (or specify the path to FFMPEG directly in the `.env` config file). Follow the "Create bot in Discord Developer Portal" section to set up the bot in Discord.

- 🎙️ Join the [Official Discord server](https://discord.gg/NVtdTka8ZT) and use the voice and `#bot-spam` channels.

## 📝 Available Discord Commands

Melodix Player supports various commands with respective aliases (if applicable). Some commands require additional parameters.

### ▶️ Playback Commands
- `!play [title|url|stream|id]` (aliases: `!p ..`, `!> ..`) — Parameters: song name, YouTube URL, audio streaming URL, history ID.
- `!skip` (aliases: `!next`, `!>>`) — Skip to the next track in the queue.
- `!pause` (alias: `!!`) — Pause playback.
- `!resume` (aliases: `!r`, `!!>`) — Resume paused playback or start playback if a track was added via `!add ..`.
- `!stop` (alias: `!x`) — Stop playback, clear the queue, and leave the voice channel.

### 📋 Queue Commands
- `!add [title|url|stream|id]` (aliases: `!a`, `!+`) — Parameters: song name, YouTube URL, audio streaming URL, history ID (same as for `!play ..`).
- `!list` (aliases: `!queue`, `!l`, `!q`) — Show the current songs queue.

### 📚 History Commands
- `!history` (aliases: `!time`, `!t`) — Show history of recently played tracks. Each track in history has a unique ID for playback/queueing.
- `!history count` (aliases: `!time count`, `!t count`) — Sort history by playback count.
- `!history duration` (aliases: `!time duration`, `!t duration`) — Sort history by track duration.

### ℹ️ Information Commands
- `!help` (aliases: `!h`, `!?`) — Show help cheatsheet.
- `!help play` — Extra information about playback commands.
- `!help queue` — Extra information about queue commands.
- `!about` (alias: `!v`) — Show version (build date) and related links.
- `whoami` — Send user-related info to the log. Needed to set up superadmin in `.env` file.

### 💾 Caching & Sideloading Commands
These commands are available only for superadmins (host server owners).
- `!curl [YouTube URL]` — Download as mp3 file for later use.
- `!cached` — Show currently cached files (from `cached` directory). Each server operates its own files.
- `!cached sync` — Synchronize manually added mp3 files to the `cached` directory.
- `!uploaded` — Show uploaded video clips in the `uploaded` directory.
- `!uploaded extract` — Extract mp3 files from video clips and store them in the `cached` directory.

### 🔧 Administration Commands
- `!register` — Enable Melodix command listening (execute once for each new Discord server).
- `!unregister` — Disable command listening.
- `melodix-prefix` — Show the current prefix (`!` by default, see `.env` file).
- `melodix-prefix-update "[new_prefix]"` — Set a custom prefix (in quotes) for a guild to avoid collisions with other bots.
- `melodix-prefix-reset` — Revert to the default prefix set in `.env` file.

### 💡 Command Usage Examples
To use the `play` command, provide a YouTube video title, URL, or history ID:
```
!play Never Gonna Give You Up
!p https://www.youtube.com/watch?v=dQw4w9WgXcQ
!> 5  (assuming 5 is an ID from `!history`)
```
For adding a song to the queue, use:
```
!add Never Gonna Give You Up
!resume
```

## 🔧 How to Set Up the Bot

### 🔗 Create a Bot in the Discord Developer Portal
To add Melodix to a Discord server, follow these steps:

1. Create an application in the [Discord Developer Portal](https://discord.com/developers/applications) and obtain the `APPLICATION_ID` (in the General section).
2. In the Bot section, enable `PRESENCE INTENT`, `SERVER MEMBERS INTENT`, and `MESSAGE CONTENT INTENT`.
3. Use the following link to authorize the bot: `discord.com/oauth2/authorize?client_id=YOUR_APPLICATION_ID&scope=bot&permissions=36727824`
   - Replace `YOUR_APPLICATION_ID` with your Bot's Application ID from step 1.
4. Select a server and click "Authorize".
5. Grant the necessary permissions for Melodix to function correctly (access to text and voice channels).

After adding the bot, build it from sources or download [compiled binaries](https://github.com/keshon/melodix-player/releases). Docker deployment instructions are available in `docker/README.md`.

### 🛠️ Building Melodix from Sources
This project is written in Go, so ensure your environment is ready. Use the provided scripts to build Melodix Player from sources:
- `bash-and-run.bat` (or `.sh` for Linux): Build the debug version and execute.
- `build-release.bat` (or `.sh` for Linux): Build the release version.
- `assemble-dist.bat`: Build the release version and assemble it as a distribution package (Windows only).

Rename `.env.example` to `.env` and store your Discord Bot Token in the `DISCORD_BOT_TOKEN` variable. Install [FFMPEG](https://ffmpeg.org/) (only recent versions are supported). If using a portable FFMPEG, specify the path in `DCA_FFMPEG_BINARY_PATH` in the `.env` file.

### 🐳 Docker Deployment
For Docker deployment, refer to `docker/README.md` for specific instructions.

## 🌐 REST API
Melodix Player provides several API routes, subject to change.

### Guild Routes
- `GET /guild/ids`: Retrieve active guild IDs.
- `GET /guild/playing`: Get info about the currently playing track in each active guild.

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

## 🆘 Support
For any questions, get support in the [Official Discord server](https://discord.gg/NVtdTka8ZT).

## 🏆 Acknowledgment
I drew inspiration from [Muzikas](https://github.com/FabijanZulj/Muzikas), a user-friendly Discord bot by Fabijan Zulj.

## 📜 License
Melodix is licensed under the [MIT License](https://opensource.org/licenses/MIT).