# ⚠️ プロジェクト終了 ⚠️

**お知らせ:** このプロジェクトはサポートされていません。開発は新しいリポジトリに移行しました：[Melodix](https://github.com/keshon/melodix)。最新の更新とサポートについては、新しいプロジェクトをご覧ください。

![# ヘッダー](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](./README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](./README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](./README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](./README_JP.md)

# 🎵 Melodix Player — Goで書かれた自己ホスティング Discord 音楽ボット

Melodix Playerは、YouTubeやオーディオストリーミングリンクからオーディオをDiscordのボイスチャンネルで再生する私の個人的なプロジェクトです。

![プレイ例](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/demo.gif)

## 🌟 機能概要

### 🎧 再生サポート
- 🎶 曲名またはYouTubeリンクで単一のトラックを追加します。
- 🎶 複数のトラックを複数のYouTubeリンク（スペース区切り）から追加します。
- 🎶 公開ユーザープレイリストからのトラック。
- 🎶 "MIX"プレイリストからのトラック。
- 📻 ストリーミングリンク（例：ラジオ局）。

### ⚙️ 追加機能
- 🌐 複数のDiscordサーバーでの操作（ギルド管理）。
- 📜 以前に再生されたトラックの履歴とソートオプションへのアクセス。
- 💾 YouTubeからトラックをmp3ファイルとしてダウンロードしてキャッシュします。
- 🎼 オーディオmp3ファイルのサイドローディング。
- 🎬 ビデオファイルをオーディオ抽出としてmp3ファイルとしてサイドロードします。
- 🔄 接続の中断時の再生自動再開サポート。
- 🛠️ REST APIサポート（現在は限定的）。

### ⚠️ 現在の制限事項
- 🚫 ボットはYouTubeストリームを再生できません。
- ⏸️ 再生自動再開サポートにより、明らかな一時停止が発生します。
- ⏩ 時々再生速度が意図よりわずかに速くなります。
- 🐞 バグが完全にないわけではありません。

## 🚀 Melodix Playerを試してみる

Melodixを試す方法は2つあります：
- 🖥️ [コンパイル済みバイナリ](https://github.com/keshon/melodix-player/releases)をダウンロードします（Windowsのみ利用可能）。システムにFFMPEGがインストールされ、グローバルPATH変数に追加されていることを確認してください（または`.env`構成ファイルで直接FFMPEGのパスを指定します）。Discordでボットを設定するには、「Discord Developer Portal」セクションに従ってボットを設定してください。

- 🎙️ [公式Discordサーバー](https://discord.gg/NVtdTka8ZT)に参加して、ボイスチャンネルと `#bot-spam` チャンネルを使用します。

## 📝 利用可能なDiscordコマンド

Melodix Playerは、それぞれのエイリアス（適用可能な場合）と共にさまざまなコマンドをサポートしています。一部のコマンドには追加のパラメータが必要です。

### ▶️ 再生コマンド
- `!play [title|url|stream|id]`（エイリアス：`!p ..`、`!> ..`）— パラメータ：曲名、YouTube URL、オーディオストリーミングURL、履歴ID。
- `!skip`（エイリアス：`!next`、`!>>`）— キュー内の次のトラックにスキップします。
- `!pause`（エイリアス：`!!`）— 再生を一時停止します。
- `!resume`（エイリアス：`!r`、`!!>`）— 一時停止した再生を再開するか、`!add ..`を介してトラックが追加された場合は再生を開始します。
- `!stop`（エイリアス：`!x`）— 再生を停止し、キューをクリアしてボイスチャンネルから退出します。

### 📋 キューコマンド
- `!add [title|url|stream|id]`（エイリアス：`!a`、`!+`）— パラメータ：曲名、YouTube URL、オーディオストリーミングURL、履歴ID（`!play ..`と同じ）。
- `!list`（エイリアス：`!queue`、`!l`、`!q`）— 現在の曲のキューを表示します。

### 📚 履歴コマンド
- `!history`（エイリアス：`!time`、`!t`）— 最近再生されたトラックの履歴を表示します。履歴の各トラックには再生/キュー待ちのための一意のIDがあります。
- `!history count`（エイリアス：`!time count`、`!t count`）— 再生回数で履歴をソートします。
- `!history duration`（エイリアス：`!time duration`、`!t duration`）— トラックの期間で履歴をソートします。

### ℹ️ 情報コマンド
- `!help`（エイリアス：`!h`、`!?`）— ヘルプチートシートを表示します。
- `!help play` — 再生コマンドに関する追加情報を表示します。
- `!help queue` — キューコマンドに関する追加情報を表示します。
- `!about`（エイリアス：`!v`）— バージョン（ビルド日付）と関連リンクを表示します。
- `whoami` — ユーザー関連の情報をログに送信します。`.env`ファイルでスーパーアドミンを設定するために必要です。

### 💾 キャッシュとサイドローディングコマンド
これらのコマンドは、スーパーアドミン（ホストサーバーの所有者）のみが利用できます。
- `!curl [YouTube URL]` — 後で使用するためにmp3ファイルとしてダウンロードします。
- `!cached` — 現在キャッシュされているファイルを表示します（`cached`ディレクトリから）。各サーバーは独自のファイルを操作します。
- `!cached sync` — 手動で追加されたmp3ファイルを`cached`ディレクトリに同期します。
- `!uploaded` — `uploaded`ディレクトリのアップロードされたビデオクリップを表示します。
- `!uploaded extract` — ビデオクリップからmp3ファイルを抽出して`cached`ディレクトリに保存します。

### 🔧 管理コマンド
- `!register` — Melodixコマンドリスニングを有効にします（新しいDiscordサーバーごとに1回実行）。
- `!unregister` — コマンドリスニングを無効にします。
- `melodix-prefix` — 現在のプレフィックスを表示します（デフォルトは `!` 、`.env`ファイルを参照）。
- `melodix-prefix-update "[new_prefix]"` — ギルドにカスタムプレフィックス（引用符内）を設定して、他のボットとの衝突を回避します。
- `melodix-prefix-reset` — `.env`ファイルで設定されたデフォルトのプレフィックスに戻します。

### 💡 コマンドの使用例
`play` コマンドを使用するには、YouTubeビデオのタイトル、URL、または履歴IDを指定します:
```
!play Never Gonna Give You Up
!p https://www.youtube.com/watch?v=dQw4w9WgXcQ
!> 5  (5 は `!history` からのIDを仮定)
```
曲をキューに追加するには、次のようにします:
```
!add Never Gonna Give You Up
!resume
```

## 🔧 ボットのセットアップ方法

### 🔗 Discord Developer Portalでボットを作成する
MelodixをDiscordサーバーに追加するには、次の手順に従ってください：

1. [Discord Developer Portal](https://discord.com/developers/applications)でアプリケーションを作成し、`APPLICATION_ID`（一般セクションにあります）を取得します。
2. ボットセクションで、`PRESENCE INTENT`、`SERVER MEMBERS INTENT`、および`MESSAGE CONTENT INTENT`を有効にします。
3. 次のリンクを使用してボットを承認します：`discord.com/oauth2/authorize?client_id=YOUR_APPLICATION_ID&scope=bot&permissions=36727824`
   - `YOUR_APPLICATION_ID` をステップ1で取得したボットのアプリケーションIDに置き換えます。
4. サーバーを選択して、「Authorize」をクリックします。
5. Melodixが正常に動作するために必要な権限を付与します（テキストチャンネルとボイスチャンネルへのアクセス）。

ボットを追加した後、ソースからビルドするか[コンパイル済みバイナリ](https://github.com/keshon/melodix-player/releases)をダウンロードします。Dockerデプロイメントの手順については、`docker/README.md` を参照してください。

### 🛠️ ソースからMelodixをビルドする
このプロジェクトはGoで書かれているため、環境が準備されていることを確認してください。提供されているスクリプトを使用して、ソースからMelodix Playerをビルドします：
- `bash-and-run.bat`（またはLinux用の`.sh`）：デバッグバージョンをビルドして実行します。
- `build-release.bat`（またはLinux用の`.sh`）：リリースバージョンをビルドします。
- `assemble-dist.bat`：リリースバージョンをビルドし、配布パッケージとして組み立てます（Windowsのみ）。

`.env.example` を `.env` にリネームし、`DISCORD_BOT_TOKEN` 変数に Discord ボットトークンを保存します。[FFMPEG](https://ffmpeg.org/) をインストールします（最新バージョンのみサポートされています）。ポータブルFFMPEGを使用する場合は、`.env`ファイルで`DCA_FFMPEG_BINARY_PATH`にパスを指定します。

### 🐳 Dockerデプロイメント
Dockerデプロイメントについては、`docker/README.md`を参照してください。

## 🌐 REST API
Melodix PlayerはいくつかのAPIルートを提供します（変更される可能性があります）。

### ギルドルート
- `GET /guild/ids`：アクティブなギルドIDを取得します。
- `GET /guild/playing`：各アクティブなギルドで現在再生中のトラックに関する情報を取得します。

### 履歴ルート
- `GET /history`：再生されたトラックの全体の履歴にアクセスします。
- `GET /history/:guild_id`：特定のギルドの再生されたトラックの履歴を取得します。

### アバタールルート
- `GET /avatar`：アバタールフォルダー内の利用可能な画像をリストします。
- `GET /avatar/random`：アバタールフォルダーからランダムな画像を取得します。

### ログルート
- `GET /log`：現在のログを表示します。
- `GET /log/clear`：ログをクリアします。
- `GET /log/download`：ログをファイルとしてダウンロードします。

## 🆘 サポート
質問がある場合は、[公式Discordサーバー](https://discord.gg/NVtdTka8ZT)でサポートを受けることができます。

## 🏆 謝辞
Fabijan Zulj氏の使いやすいDiscordボット「Muzikas」からインスピレーションを受けました。

## 📜 ライセンス
Melodixは[MITライセンス](https://opensource.org/licenses/MIT)の下でライセンスされています。