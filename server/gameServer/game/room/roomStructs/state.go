package roomStructs

type RoomState string

// estados possíveis da sala
const (
	StateWaitAction           RoomState = "WAIT_ACTION"
	StateWaitContest          RoomState = "WAIT_CONTEST"
	StateWaitPunishment       RoomState = "WAIT_PUNISHMENT"
	StateWaitKamikazeDecision RoomState = "WAIT_KAMIKAZE"
)
