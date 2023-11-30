# Melodix directory

## player.go
`Player` - Player that streams given song to specific Discord channel. Minimal walkman functionality.

## discord.go
`Discord` - Discord commands that overlay on top of a player. There is a dedicated instance per each Discord server.

## restful.go
`Rest` - Rest API commands to do certain actions using HTTP requests.

## youtube.go
`Youtube` - Methods to handle track parsing from Youtube.

## history.go
`History` - List of played track for particual instance for easy replaying.

## utility.go
Various independent little functions to fulfil needs of the above.