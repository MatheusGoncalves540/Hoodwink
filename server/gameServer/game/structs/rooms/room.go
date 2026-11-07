package rooms

import (
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
)

type Room struct {
	ID                        string                                                `json:"id"`
	Rules                     structs.Rules                                         `json:"rules"`
	TimeoutType               string                                                `json:"timeoutType"`
	Name                      string                                                `json:"name"`
	Password                  string                                                `json:"password" validate:"max=24"`
	MaxPlayers                int                                                   `json:"maxPlayers"`
	CustomMatch               bool                                                  `json:"customMatch"`
	Turn                      int                                                   `json:"turn"`
	Tax                       int                                                   `json:"tax"`
	DoubledCardValues         map[structs.TypePlayerPlays]structs.DoubledCardValues `json:"doubledCardValues"`
	Players                   map[string]*players.Player                            `json:"players"`
	Deck                      []structs.CardName                                    `json:"deck"`
	CurrentPlayer             string                                                `json:"currentPlayer"`
	GameEvent                 *structs.GameEvent                                    `json:"gameEvent"`
	PendingEffects            []structs.EffectDTO                                   `json:"pendingEffects"`
	PendingPresentationEvents []structs.PresentationEventDTO                        `json:"pendingPresentationEvents"`
	GameOver                  bool                                                  `json:"gameOver"`
	WinnerID                  string                                                `json:"winnerId,omitempty"`
	StartTime                 time.Time                                             `json:"startTime"`
	Created                   time.Time                                             `json:"created"`
	Chat                      []structs.ChatMessage                                 `json:"chat"`

	// MatchID é o UUID v7 público da partida (o plano 06 usa).
	MatchID string `json:"matchId"`
	// Seed é a semente de 128 bits; serializada no JSON da sala (Redis) mas nunca
	// vai ao cliente/log (String/GoString/LogValue escondem o valor).
	Seed Seed `json:"seed"`
	// Rng guarda o estado do gerador PCG da sala; serializado no JSON (Redis).
	Rng RngState `json:"rng"`
	// NextMoveTime é o contador de ordem de turno: MoveTime nunca repete.
	NextMoveTime int `json:"nextMoveTime"`

	// Connected é o conjunto ordenado de ids dos jogadores conectados à sala,
	// preenchido pelo WithRoom a partir do registro no Redis antes do fn. O motor
	// nunca consulta o Redis para isso: lê este campo.
	Connected []string `json:"connected,omitempty"`
	// NonReplayable marca a partida como não replayável (ex.: comando de debug).
	NonReplayable bool `json:"nonReplayable,omitempty"`

	// lock é o lock distribuído ativo durante uma chamada do WithRoom. Não é
	// serializado (o WithRoom preenche e o JSON/Clone ignora campos não exportados).
	lock RoomLock

	// now é o relógio fixo da sala durante uma chamada do WithRoom. Não é
	// serializado; o WithRoom preenche e solta via SetNow/ClearNow.
	now *time.Time
}

// AttachLock associa o lock ativo a esta sala (setado pelo WithRoom durante o fn).
func (r *Room) AttachLock(l RoomLock) {
	r.lock = l
}

// Lock retorna o lock distribuído ativo desta sala, ou nil quando fora do WithRoom.
func (r *Room) Lock() RoomLock {
	return r.lock
}

type PublicRoomForUpdates struct {
	ID string `json:"id"`
	// MatchID é público: o cliente usa para o relato de bug e o replay.
	MatchID           string                                                `json:"matchId"`
	Name              string                                                `json:"name"`
	MaxPlayers        int                                                   `json:"maxPlayers"`
	Turn              int                                                   `json:"turn"`
	Tax               int                                                   `json:"tax"`
	DoubledCardValues map[structs.TypePlayerPlays]structs.DoubledCardValues `json:"doubledCardValues"`
	Players           map[string]players.PublicPlayerForUpdates             `json:"players"`
	Deck              int                                                   `json:"deck"`
	CurrentPlayer     string                                                `json:"currentPlayer"`
	GameEvent         *structs.GameEvent                                    `json:"gameEvent"`
	// Sem PendingEffects: a fila interna do motor não é pública (alvos e reações ainda não anunciados).
	GameOver  bool                  `json:"gameOver"`
	WinnerID  string                `json:"winnerId,omitempty"`
	StartTime time.Time             `json:"startTime"`
	Chat      []structs.ChatMessage `json:"chat"`
}
