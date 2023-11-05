# Melodix directory

## player.go
`MelodixPlayer` - Player that streams given song to specific Discord channel. Minimal walkman functionality.

## discord.go
`DiscordMelodix` - Discord commands that overlay on top of a player. There is a dedicated instance per each Discord server.

## restful.go
`RestfulMelodix` - Rest API commands to do certain actions using HTTP requests.

## song.go
Various functions with youtube song parsing methods.

## history.go
`MelodixHistory` - List of played track for particual instance for easy replaying.

## utility.go
Various independent little functions to fulfil needs of the above.