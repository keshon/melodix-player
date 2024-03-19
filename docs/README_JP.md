![# Header](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/banner-readme.png)

[![Español](https://img.shields.io/badge/Español-README-blue)](./README_ES.md) [![Français](https://img.shields.io/badge/Français-README-blue)](./README_FR.md) [![中文](https://img.shields.io/badge/中文-README-blue)](./README_CN.md) [![日本語](https://img.shields.io/badge/日本語-README-blue)](./README_JP.md)

# Melodix Player

Melodix Playerは、接続エラーが発生しても最善を尽くすDiscord音楽ボットです。

## 機能概要

このボットは、使いやすいが強力な音楽プレーヤーを目指しています。主な目標は次のとおりです。

- YouTubeからの単一/複数のトラックまたはプレイリストの再生（タイトルまたはURLで追加）。
- URL経由で追加されたラジオストリームの再生。
- 再生回数や再生時間に対するソートオプション付きで、以前に再生されたトラックの履歴へのアクセス。
- ネットワークの障害による再生の中断に対処 - Melodixは再生を再開しようとします。
- Discordコマンド以外のさまざまなタスクを実行するための露出されたRest API。
- 複数のDiscordサーバーでの動作。

![再生の例](https://raw.githubusercontent.com/keshon/melodix-player/master/assets/demo.gif)

## バイナリのダウンロード

バイナリ（Windowsのみ）は[リリースページ](https://github.com/keshon/melodix-player/releases)で利用可能です。最新バージョンはソースからビルドすることが推奨されています。

## Discordコマンド

Melodix Playerは、音楽の再生を制御するためのさまざまなコマンドとそれに対応するエイリアスをサポートしています。一部のコマンドには追加のパラメータが必要です。

**コマンドとエイリアス**：
- `play` (`p`, `>`) — パラメータ：YouTubeのビデオURL、履歴ID、トラックのタイトル、または有効なストリームリンク
- `skip` (`next`, `ff`, `>>`)
- `pause` (`!`)
- `resume` (`r`,`!>`)
- `stop` (`x`)
- `add` (`a`, `+`) — パラメータ：YouTubeのビデオURLまたは履歴ID、トラックのタイトル、または有効なストリームリンク
- `list` (`queue`, `l`, `q`)
- `history` (`time`, `t`) — パラメータ： `duration` または `count`
- `help` (`h`, `?`)
- `about` (`v`)
- `register`
- `unregister`

コマンドはデフォルトで `!` でプレフィックスを付ける必要があります。たとえば、`!play`、`!>>`などです。

### 例
`play` コマンドを使用するには、YouTubeのビデオのタイトル、URL、または履歴IDをパラメータとして指定します。例：
`!play Never Gonna Give You Up` 
または 
`!p https://www.youtube.com/watch?v=dQw4w9WgXcQ` 
または 
`!> 5`（`5`が履歴から見えるIDであると仮定： `!history`）

曲をキューに追加するには、同様のアプローチを取ります：
`!add Never Gonna Give You Up` 
`!resume`（再生を開始する）

## Discordサーバーにボットを追加する

MelodixをDiscordサーバーに追加するには：

1. [Discord Developer Portal](https://discord.com/developers/applications)でボットを作成し、Botの`CLIENT_ID`を取得します。
2. 次のリンクを使用します： `discord.com/oauth2/authorize?client_id=YOUR_CLIENT_ID_HERE&scope=bot&permissions=36727824`
   - `YOUR_CLIENT_ID_HERE` をボットのクライアントIDに置き換えます（ステップ1で取得）。
3. Discordの認証ページがブラウザで開き、サーバーを選択できます。
4. Melodixを追加したいサーバーを選択し、「Authorize」をクリックします。
5. Melodixに正常に機能するために必要な権限を付与します。

ボットが追加されたら、実際のボットの構築に進んでください。

## ソースからビルドする

このプロジェクトはGo言語で書かれており、*サーバー*または*ローカル*プログラムとして実行できます。

**ローカルの使用**
Melodix Playerをソースからビルドするためのいくつかのスクリプトが用意されています：
- `bash-and-run.bat`（またはLinux用の`.sh`）：デバッグバージョンをビルドして実行します。
- `build-release.bat`（またはLinux用の`.sh`）：リリースバージョンをビルドします。
- `assemble-dist.bat`：リリースバージョンをビルドし、配布パッケージとして組み立てます（Windowsのみ、プロセス中にUPXパッケージャがダウンロードされます）。

ローカルで使用する場合は、これらのスクリプトを操作するためのオペレーティングシステムごとに実行し、`.env.example` を `.env` にリネームし、`DISCORD_BOT_TOKEN` 変数にDiscordボットトークンを格納します。 [FFMPEG](https://ffmpeg.org/) をインストールします（最新バージョンのみサポートされます）。 FFMPEGのインストールがポータブルな場合は、`DCA_FFMPEG_BINARY_PATH` 変数にパスを指定します。

**サーバーの使用**
Docker環境でボットをビルドしてデプロイするには、`docker/README.md` を参照してください。

バイナリファイルがビルドされ、`.env` ファイルが記入され、ボットがサーバーに追加されたら、Melodixは操作の準備ができています。

## APIアクセスとルート

Melodix Playerはさまざまな機能に対するさまざまなルートを提供しています：

### ギルドルート

- `GET /guild/ids`：アクティブなギルドIDを取得します。
- `GET /guild/playing`：各アクティブなギルドで現在再生中のトラックに関する情報を取得します。

### 履歴ルート

- `GET /history`：再生されたトラックの全体の履歴にアクセスします。
- `GET /history/:guild_id`：特定のギルドの再生されたトラックの履歴を取得します。

### アバタールルート

- `GET /avatar`：アバターフォルダ内の利用可能な画像をリスト表示します。
- `GET /avatar/random`：アバターフォルダからランダムな画像を取得します。

### ログルート

- `GET /log`：現在のログを表示します。
- `GET

 /log/clear`：ログをクリアします。
- `GET /log/download`：ログをファイルとしてダウンロードします。

## サポートを受ける場所

質問があれば、[Discordサーバー](https://discord.gg/NVtdTka8ZT)で質問してサポートを受けることができます。何のコミュニティもないことを心に留めておいてください。

## 謝辞

[Fabijan Zulj](https://github.com/FabijanZulj)によって作成された使いやすいDiscordボット[Muzikas](https://github.com/FabijanZulj/Muzikas)からインスピレーションを得ました。

Melodixの開発の結果、新しいプロジェクトが誕生しました：[Discord Bot Boilerplate](https://github.com/keshon/discord-bot-boilerplate) — Discordボットを構築するためのフレームワーク。

## ライセンス

Melodixは[MITライセンス](https://opensource.org/licenses/MIT)のもとで提供されています。