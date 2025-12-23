package roomStructs

type RoomState string

// estados possíveis da sala
const (
	StateWaitAction           RoomState = "WAIT_ACTION"
	StateWaitTurn             RoomState = "WAIT_TURN"
	StateWaitContest          RoomState = "WAIT_CONTEST"
	StateWaitPunishment       RoomState = "WAIT_PUNISHMENT"
	StateWaitKamikazeDecision RoomState = "WAIT_KAMIKAZE"
)
