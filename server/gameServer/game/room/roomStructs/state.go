package roomStructs

// Definindo os estados possíveis da sala
const (
	WaitingGameStart        = "waitingGameStart"
	WaitingFirstAction      = "waitingFirstAction"
	WaitingAction           = "waitingAction"
	WaitingContest          = "waitingContest"
	ResolvingContest        = "resolvingContest"
	WaitingKamikazeResponse = "waitingKamikazeResponse"
	FinalizingAction        = "finalizingAction"
	TurnFinished            = "turnFinished"
)
