![# Header](https://github.com/keshon/melodix-player/blob/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](/docs/README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](/docs/README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](/docs/README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](/docs/README_JP.md)

# Melodix Player

Melodix Player是一个Discord音乐机器人，即使在存在连接错误的情况下，也会尽力而为。

## 功能概述

该机器人旨在成为一个易于使用但功能强大的音乐播放器。其主要目标包括：

- 从YouTube播放单个/多个曲目或播放列表，通过标题或URL添加。
- 通过URL播放添加的广播流。
- 访问先前播放曲目的历史记录，可按播放次数或持续时间进行排序。
- 处理由于网络故障而中断的播放 - Melodix将尝试恢复播放。
- 提供暴露的Rest API以执行在Discord命令之外的各种任务。
- 跨多个Discord服务器运行。

![播放示例](https://github.com/keshon/melodix-player/blob/master/assets/playing.jpg)

## 下载二进制文件

二进制文件（仅限Windows）可在[发布页面](https://github.com/keshon/melodix-player/releases)上找到。建议从源代码构建二进制文件以获取最新版本。

## Discord命令

Melodix Player支持各种命令及其相应的别名来控制音乐播放。一些命令需要额外的参数：

**命令和别名**：
- `play` (`p`, `>`) — 参数：YouTube视频URL、历史ID、曲目标题或有效的流链接
- `skip` (`next`, `ff`, `>>`)
- `pause` (`!`)
- `resume` (`r`,`!>`)
- `stop` (`x`)
- `add` (`a`, `+`) — 参数：YouTube视频URL或历史ID、曲目标题或有效的流链接
- `list` (`queue`, `l`, `q`)
- `history` (`time`, `t`) — 参数：`duration`或`count`
- `help` (`h`, `?`)
- `about` (`v`)
- `register`
- `unregister`

默认情况下，命令应以 `!` 为前缀。例如，`!play`，`!>>`等。

### 示例
要使用 `play` 命令，请提供YouTube视频标题、URL或历史ID作为参数，例如：
`!play Never Gonna Give You Up` 
或 
`!p https://www.youtube.com/watch?v=dQw4w9WgXcQ` 
或 
`!> 5`（假设 `5` 是可以从历史记录中看到的ID：`!history`）

要将歌曲添加到队列中，请使用类似的方法：
`!add Never Gonna Give You Up` 
`!resume`（开始播放）

## 将机器人添加到Discord服务器

要将Melodix添加到您的Discord服务器：

1. 在[Discord Developer Portal](https://discord.com/developers/applications)上创建一个机器人并获取Bot的`CLIENT_ID`。
2. 使用以下链接：`discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID_HERE&scope=bot&permissions=36727824`
   - 用步骤1中获取的Bot的客户端ID替换`YOUR_CLIENT_ID_HERE`。
3. Discord授权页面将在您的浏览器中打开，允许您选择一个服务器。
4. 选择要添加Melodix的服务器，然后点击“Authorize”。
5. 如果提示，请完成reCAPTCHA验证。
6. 授予Melodix正常运行所需的权限。
7. 点击“Authorize”将Melodix添加到您的服务器。

一旦机器人被添加，就可以继续进行实际的机器人构建。

## API访问和路由

Melodix Player为不同功能提供了各种路由：

### 公会路由

- `GET /guild/ids`：检索活动公会ID。
- `GET /guild/playing`：获取每个活动公会中当前正在播放的曲目的信息。

### 历史路由

- `GET /history`：访问播放曲目的整体历史记录。
- `GET /history/:guild_id`：获取特定公会的播放曲目历史记录。

### 头像路由

- `GET /avatar`：列出头像文件夹中可用的图像。
- `GET /avatar/random`：从头像文件夹获取随机图像。

### 日志路由

- `GET /log`：显示当前日志。
- `GET /log/clear`：清除日志。
- `GET /log/download`：将日志下载为文件。

## 从源代码构建

此项目使用Go语言编写，可在*服务器*或*本地*程序上运行。

**本地使用**
提供了几个脚本用于从源代码构建Melodix Player：
- `bash-and-run.bat`（或Linux用的`.sh`）：构建调试版本并执行。
- `build-release.bat`（或Linux用的`.sh`）：构建发布版本。
- `assemble-dist.bat`：构建发布版本并将其组装为分发包（仅限Windows，在此过程中将下载UPX打包器）。

对于本地使用，请针对您的操作系统运行这些脚本，并将`.env.example`重命名为`.env`，将您的Discord Bot Token存储在`DISCORD_BOT_TOKEN`变量中。安装[FFMPEG](https://ffmpeg.org/)（仅支持最新版本）。如果您的FFMPEG安装是便携式的，请在`DCA_FFMPEG_BINARY_PATH`变量中指定路径。

**服务器使用**
要在Docker环境中构建和部署机器人，请参阅`docker/README.md`以获取具体说明。

一旦构建了二进制文件，填充了`.env`文件，并将Bot添加到服务器，Melodix就准备好运行了。

## 获取支持的地方

如果有任何问题，您可以在我的[Discord服务器](https://discord.gg/NV

tdTka8ZT)中问我以获取支持。请注意，几乎没有社区 - 只有我。

## 致谢

我从[Fabijan Zulj](https://github.com/FabijanZulj)创建的用户友好的Discord机器人[Muzikas](https://github.com/FabijanZulj/Muzikas)中汲取了灵感。

由于Melodix的开发，诞生了一个新项目：[Discord Bot Boilerplate](https://github.com/keshon/discord-bot-boilerplate) —— 用于构建Discord机器人的框架。

## 许可证

Melodix根据[MIT许可证](https://opensource.org/licenses/MIT)获得许可。