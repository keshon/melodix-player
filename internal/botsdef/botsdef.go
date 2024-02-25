package botsdef

type Discord interface {
	Start(guildID string)
	Stop()
}
