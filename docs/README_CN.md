![# Header](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](./README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](./README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](./README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](./README_JP.md)

# 🎵 Melodix Player — 用 Go 编写的自托管 Discord 音乐机器人

Melodix Player 是我的宠物项目，它可以将 YouTube 和音频流链接中的音频播放到 Discord 语音频道。

![Playing Example](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/demo.gif)

## 🌟 功能概述

### 🎧 播放支持
- 🎶 通过歌曲名称或 YouTube 链接添加单曲。
- 🎶 通过多个 YouTube 链接（以空格分隔）添加多首歌曲。
- 🎶 从公共用户播放列表中添加歌曲。
- 🎶 从“混音”播放列表中添加歌曲。
- 📻 流媒体链接（例如，广播电台）。

### ⚙️ 其他功能
- 🌐 支持跨多个 Discord 服务器（公会管理）。
- 📜 访问带有排序选项的已播放曲目历史记录。
- 💾 下载 YouTube 的曲目为 mp3 文件以进行缓存。
- 🎼 加载音频 mp3 文件。
- 🎬 加载视频文件并提取音频为 mp3 文件。
- 🔄 支持连接中断时的播放自动恢复。
- 🛠️ 支持 REST API（目前有限）。

### ⚠️ 目前的限制
- 🚫 机器人不能播放 YouTube 流。
- ⏸️ 播放自动恢复支持会产生明显的暂停。
- ⏩ 有时播放速度会比预期的稍快。
- 🐞 不是没有 bug 的。

## 🚀 试用 Melodix Player

您可以通过两种方式试用 Melodix：
- 🖥️ 下载[编译好的二进制文件](https://github.com/keshon/melodix-player/releases)（仅适用于 Windows）。确保您的系统已安装 FFMPEG 并将其添加到全局 PATH 变量中（或直接在 `.env` 配置文件中指定 FFMPEG 的路径）。按照“在 Discord 开发者门户中创建机器人”部分设置 Discord 中的机器人。

- 🎙️ 加入[官方 Discord 服务器](https://discord.gg/NVtdTka8ZT)并使用语音和 `#bot-spam` 频道。

## 📝 可用的 Discord 命令

Melodix Player 支持各种命令及其对应的别名（如果适用）。某些命令需要额外的参数。

### ▶️ 播放命令
- `!play [title|url|stream|id]`（别名：`!p ..`，`!> ..`） — 参数：歌曲名称、YouTube URL、音频流 URL、历史记录 ID。
- `!skip`（别名：`!next`，`!>>`） — 跳过队列中的下一首歌曲。
- `!pause`（别名：`!!`） — 暂停播放。
- `!resume`（别名：`!r`，`!!>`） — 恢复暂停的播放或如果通过 `!add ..` 添加了曲目则开始播放。
- `!stop`（别名：`!x`） — 停止播放，清空队列并离开语音频道。

### 📋 队列命令
- `!add [title|url|stream|id]`（别名：`!a`，`!+`） — 参数：歌曲名称、YouTube URL、音频流 URL、历史记录 ID（与 `!play ..` 相同）。
- `!list`（别名：`!queue`，`!l`，`!q`） — 显示当前的歌曲队列。

### 📚 历史命令
- `!history`（别名：`!time`，`!t`） — 显示最近播放的曲目历史记录。每个历史记录中的曲目都有一个唯一的 ID 可用于播放/排队。
- `!history count`（别名：`!time count`，`!t count`） — 按播放次数排序历史记录。
- `!history duration`（别名：`!time duration`，`!t duration`） — 按曲目时长排序历史记录。

### ℹ️ 信息命令
- `!help`（别名：`!h`，`!?`） — 显示帮助速查表。
- `!help play` — 额外的播放命令信息。
- `!help queue` — 额外的队列命令信息。
- `!about`（别名：`!v`） — 显示版本（构建日期）和相关链接。
- `whoami` — 将用户相关信息发送到日志。需要在 `.env` 文件中设置超级管理员。

### 💾 缓存和加载命令
这些命令仅对超级管理员（主机服务器所有者）可用。
- `!curl [YouTube URL]` — 下载为 mp3 文件以供以后使用。
- `!cached` — 显示当前缓存的文件（在 `cached` 目录中）。每个服务器都有自己的文件。
- `!cached sync` — 同步手动添加的 mp3 文件到 `cached` 目录。
- `!uploaded` — 显示 `uploaded` 目录中的上传视频剪辑。
- `!uploaded extract` — 从视频剪辑中提取 mp3 文件并将其存储在 `cached` 目录中。

### 🔧 管理命令
- `!register` — 启用 Melodix 命令监听（每个新的 Discord 服务器执行一次）。
- `!unregister` — 禁用命令监听。
- `melodix-prefix` — 显示当前的前缀（默认是 `!`，见 `.env` 文件）。
- `melodix-prefix-update "[new_prefix]"` — 为公会设置自定义前缀以避免与其他机器人的冲突。
- `melodix-prefix-reset` — 恢复为 `.env` 文件中设置的默认前缀。

### 💡 命令使用示例
要使用 `play` 命令，请提供 YouTube 视频标题、URL 或历史记录 ID：
```
!play Never Gonna Give You Up
!p https://www.youtube.com/watch?v=dQw4w9WgXcQ
!> 5  （假设 5 是 `!history` 中的 ID）
```
要将歌曲添加到队列，请使用：
```
!add Never Gonna Give You Up
!resume
```

## 🔧 如何设置机器人

### 🔗 在 Discord 开发者门户中创建机器人
要将 Melodix 添加到 Discord 服务器，请按照以下步骤操作：

1. 在[Discord 开发者门户](https://discord.com/developers/applications)中创建一个应用程序，并获取 `APPLICATION_ID`（在常规部分）。
2. 在机器人部分，启用 `PRESENCE INTENT`，`SERVER MEMBERS INTENT` 和 `MESSAGE CONTENT INTENT`。
3. 使用以下链接授权机器人：`discord.com/oauth2/authorize?client_id=YOUR_APPLICATION_ID&scope=bot&permissions=36727824`
   - 将 `YOUR_APPLICATION_ID` 替换为步骤 1 中的机器人的应用程序 ID。
4. 选择一个服务器并点击“授权”。
5. 授予 Melodix 正常运行所需的权限（访问文本和语音频道）。

添加机器人后，从源代码构建或下载[编译好的二进制文件](https://github.com/keshon/melodix-player/releases)。Docker 部署说明请参见 `docker/README.md`。

### 🛠️ 从源代码构建 Melodix
该项目是用 Go 编写的，所以确保您的环境已准备好。使用提供的脚本从源代码构建 Melodix Player：
- `bash-and-run.bat`（或 Linux 的 `.sh`）：构建调试版本并执行。
- `build-release.bat`（或 Linux 的 `.sh`）：构建发布版本。
- `assemble-dist.bat`：构建发布版本并将其组装为发行包（仅适用于 Windows）。

将 `.env.example` 重命名为 `.env` 并将您的 Discord 机器人令牌存储在 `DISCORD_BOT_TOKEN` 变量中。安装 [FFMPEG](https://ffmpeg.org/)（仅支持最新版本）。如果使用便携版 FFMPEG，请在 `.env` 文件中指定路径 `DCA_FFMPEG_BINARY_PATH`。

### 🐳 Docker 部署
有关 Docker 部署的具体说明，请参见 `

docker/README.md`。

## 🌐 REST API
Melodix Player 提供多个 API 路由，但可能会有变动。

### 公会路由
- `GET /guild/ids`：检索活动公会 ID。
- `GET /guild/playing`：获取每个活动公会当前播放的曲目信息。

### 历史路由
- `GET /history`：访问已播放曲目的整体历史记录。
- `GET /history/:guild_id`：获取特定公会已播放曲目的历史记录。

### 头像路由
- `GET /avatar`：列出头像文件夹中的可用图像。
- `GET /avatar/random`：从头像文件夹中获取随机图像。

### 日志路由
- `GET /log`：显示当前日志。
- `GET /log/clear`：清除日志。
- `GET /log/download`：下载日志文件。

## 🆘 支持
如有任何问题，请在[官方 Discord 服务器](https://discord.gg/NVtdTka8ZT)获取支持。

## 🏆 鸣谢
我从 [Muzikas](https://github.com/FabijanZulj/Muzikas) 获得灵感，这是一款由 Fabijan Zulj 开发的用户友好的 Discord 机器人。

## 📜 许可证
Melodix 采用 [MIT 许可证](https://opensource.org/licenses/MIT)。