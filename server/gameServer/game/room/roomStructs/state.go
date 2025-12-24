package roomStructs

type RoomState string

// estados possíveis da sala
const (
	StateWaitAction         RoomState = "WAIT_ACTION"
	StateWaitingFirstAction RoomState = "WAITING_FIRST_ACTION"
	StateWaitContest        RoomState = "WAIT_CONTEST"
)
