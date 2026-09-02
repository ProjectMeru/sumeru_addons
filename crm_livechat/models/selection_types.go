package models

type LivechatState string

const (
	LivechatStateNew     LivechatState = "new"
	LivechatStateActive  LivechatState = "active"
	LivechatStateDone    LivechatState = "done"
)
