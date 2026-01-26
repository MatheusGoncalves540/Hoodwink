package rooms

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/rules"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/players"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/redisFuncs/playerRedis"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/utils"
	"github.com/redis/go-redis/v9"
)

type Room struct {
	ID                string                                                `json:"id"`
	Rules             structs.Rules                                         `json:"rules"`
	TimeoutType       string                                                `json:"timeoutType"`
	Name              string                                                `json:"name"`
	Password          string                                                `json:"password" validate:"max=24"`
	MaxPlayers        int                                                   `json:"maxPlayers"`
	CustomMatch       bool                                                  `json:"customMatch"`
	Turn              int                                                   `json:"turn"`
	Tax               int                                                   `json:"tax"`
	DoubledCardValues map[structs.TypePlayerPlays]structs.DoubledCardValues `json:"doubledCardValues"`
	Players           map[string]*players.Player                            `json:"players"`
	Deck              []string                                              `json:"deck"`
	CurrentPlayer     string                                                `json:"currentPlayer"`
	GameEvent         *structs.GameEvent                                    `json:"gameEvent"`
	PendingEffects    []structs.EffectDTO                                   `json:"pendingEffects"`
	GameOver          bool                                                  `json:"gameOver"`
	StartTime         time.Time                                             `json:"startTime"`
	Created           time.Time                                             `json:"created"`
}

type PublicRoomForUpdates struct {
	ID                string                                                `json:"id"`
	Turn              int                                                   `json:"turn"`
	Tax               int                                                   `json:"tax"`
	DoubledCardValues map[structs.TypePlayerPlays]structs.DoubledCardValues `json:"doubledCardValues"`
	Players           map[string]players.PublicPlayerForUpdates             `json:"players"`
	Deck              []string                                              `json:"deck"`
	CurrentPlayer     string                                                `json:"currentPlayer"`
	GameEvent         *structs.GameEvent                                    `json:"gameEvent"`
	PendingEffects    []structs.EffectDTO                                   `json:"pendingEffects"`
	GameOver          bool                                                  `json:"gameOver"`
	StartTime         time.Time                                             `json:"startTime"`
}

// GetPlayer retorna o ponteiro do jogador pela playerId e um bool indicando se existe
func (r *Room) GetPlayer(playerId string) (*players.Player, error) {
	player, exists := r.Players[playerId]
	if !exists {
		return nil, fmt.Errorf("jogador alvo não encontrado: %s", playerId)
	}
	return player, nil
}

// GetCardRules retorna as regras da carta específica para a sala
func (r *Room) GetCardRules(registryRules *rules.Registry, card string) (*rules.CardRules, error) {
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return nil, err
	}
	cardRules := rules.CardRules(gameRules.Cards[card])
	return &cardRules, nil
}

// GetGeneralRules retorna as regras gerais do jogo para a sala
func (r *Room) GetGeneralRules(registryRules *rules.Registry) (*rules.Generals, error) {
	// Obtém as regras do jogo para a sala
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return nil, err
	}
	return &gameRules.Generals, nil
}

// GetTimeoutDuration retorna a duração do timeout específico
func (r *Room) GetTimeoutDuration(registryRules *rules.Registry, timeoutField string) (time.Duration, error) {
	// Obtém as regras do jogo para a sala
	gameRules, err := registryRules.Get(string(r.Rules))
	if err != nil {
		return 0, err
	}

	// Obtém a configuração de timeout para o tipo da sala
	timeoutConfig, exists := gameRules.Timeouts[r.TimeoutType]
	if !exists {
		return 0, fmt.Errorf("tipo de timeout '%s' não encontrado", r.TimeoutType)
	}

	// Retorna o valor do campo via função helper
	value, err := timeoutConfig.Get(timeoutField)
	if err != nil {
		return 0, fmt.Errorf("erro ao obter timeoutField '%s': %w", timeoutField, err)
	}
	return time.Duration(value), nil
}

// SaveRoomWithTTL salva o estado da sala no Redis com TTL específico
func (r *Room) SaveRoomWithTTL(ctx context.Context, rdb *redis.Client, ttl time.Duration) error {
	// Pega o ID da instância atual
	instanceID := utils.GetInstanceID()
	// Tenta adquirir o lock da sala
	ok, err := r.AcquireRoomLock(ctx, rdb, instanceID, 5*time.Second)
	if err != nil {
		// Retorna erro se falhar ao tentar lock
		utils.LogError(err)
		return err
	}
	if !ok {
		// Retorna erro se outra instância já está modificando
		return fmt.Errorf("não foi possível adquirir lock para a sala %s", r.ID)
	}
	// Libera o lock ao final da função
	defer r.ReleaseRoomLock(ctx, rdb, instanceID)

	// Serializa a sala para JSON
	data, err := json.Marshal(r)
	if err != nil {
		// Retorna erro se não conseguir serializar
		return err
	}
	// Salva o JSON no Redis com TTL específico
	err = rdb.Set(ctx, "room:"+r.ID, data, ttl).Err()
	if err != nil {
		return err
	}

	return nil
}

// SaveRoom salva o estado da sala no Redis de forma segura usando lock distribuído.
func (r *Room) SaveRoom(ctx context.Context, rdb *redis.Client) error {
	roomAfkTtlLimit := utils.MustEnvInt("ROOM_AFK_TTL_LIMIT", 180) // 3 minutos padrão
	err := r.SaveRoomWithTTL(ctx, rdb, time.Duration(roomAfkTtlLimit)*time.Second)
	if err != nil {
		return err
	}
	return nil
}

// AddPlayerInRoom adiciona um jogador à sala caso ele ainda não exista na estrutura da sala
func (r *Room) AddPlayerInRoom(ctx context.Context, rdb *redis.Client, playerId string, username string) error {
	// Verifica se o jogador já existe
	player, err := r.GetPlayer(playerId)
	if err == nil && player != nil {
		// Jogador já existe, nada a fazer
		return nil
	}

	// Adiciona o jogador ao mapa
	r.Players[playerId] = &players.Player{
		Id:   playerId,
		Name: username,
		Cards: []structs.Card{
			{Name: structs.CardKamikaze, Index: 0, Protected: false, Dead: false},
			{Name: structs.CardKamikaze, Index: 1, Protected: false, Dead: false},
		}, // TODO remover valor fixo de teste
		Coins: 10, // Valor inicial padrão do jogo
		Alive: true,
	}

	// Salva a sala atualizada
	err = r.SaveRoom(ctx, rdb)
	return nil
}

// AcquireRoomLock tenta adquirir ou renovar um lock distribuído no redis para uma sala.
func (r *Room) AcquireRoomLock(ctx context.Context, rdb *redis.Client, instanceID string, ttl time.Duration) (bool, error) {
	key := "lock:room:" + r.ID

	// 1. Tenta adquirir o lock normalmente
	ok, err := rdb.SetNX(ctx, key, instanceID, ttl).Result()
	if err != nil {
		return false, err
	}

	if ok {
		// Lock adquirido com sucesso
		return true, nil
	}

	// 2. Lock já existe — verificar se é da mesma instância
	val, err := rdb.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// A chave sumiu entre o SetNX e o Get (condição rara)
			return false, nil
		}
		return false, err
	}

	if val != instanceID {
		// Lock pertence a outra instância
		return false, nil
	}

	// 3. Lock é da mesma instância → renovar TTL
	_, err = rdb.Expire(ctx, key, ttl).Result()
	if err != nil {
		return false, err
	}

	return true, nil
}

// ReleaseRoomLock remove o lock redis da sala se ainda pertencer à instância atual.
func (r *Room) ReleaseRoomLock(ctx context.Context, rdb *redis.Client, instanceID string) error {
	// Busca o valor atual do lock
	val, err := rdb.Get(ctx, "lock:room:"+r.ID).Result()
	if err == nil && val == instanceID {
		// Só remove se o lock for da instância
		return rdb.Del(ctx, "lock:room:"+r.ID).Err()
	}
	// Não remove se não for o dono ou se não existir
	return nil
}

// RemovePlayerFromRoom remove completamente um jogador da estrutura da sala
func (r *Room) RemovePlayerFromRoom(ctx context.Context, rdb *redis.Client, playerId string) error {
	// Verifica se o jogador existe
	player, err := r.GetPlayer(playerId)
	if err != nil {
		return err
	}
	if player == nil {
		return fmt.Errorf("jogador %s não encontrado na sala %s", playerId, r.ID)
	}

	// Remove o jogador do mapa
	delete(r.Players, playerId)

	// Salva a sala atualizada
	return r.SaveRoom(ctx, rdb)
}

// Publica mensagem para todos os players de uma sala
func (r *Room) BroadcastMsgToRoom(ctx context.Context, rdb *redis.Client, message any) error {
	if pubMsg, err := json.Marshal(message); err == nil {
		return rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// SendUpdatedRoomData publica nova versão da sala para todos os players ou para players específicos.
// Se confidencialRoomData for nil OU confidencialPlayerIds estiver vazio, envia para todos (toAll: true).
// Caso contrário, envia apenas para os playerIds especificados usando confidencialRoomData.
func (r *Room) SendUpdatedRoomData(ctx context.Context, rdb *redis.Client, confidencialRoomData *Room, confidencialPlayerIds []string) error {
	// Verifica se é broadcast para todos ou seletivo
	toAll := confidencialRoomData == nil || len(confidencialPlayerIds) == 0

	if toAll {
		// Broadcast para todos - usa roomData pública normal
		if utils.IsUnsafeDebugMode() {
			return r.PublishUnsafeRoomData(ctx, rdb, true, nil)
		} else {
			return r.PublishRoomUpdate(ctx, rdb, true, nil)
		}
	} else {
		// Broadcast para todos - usa roomData pública normal
		// depois envia Broadcast seletivo - usa confidencialRoomData
		if utils.IsUnsafeDebugMode() {
			r.PublishUnsafeRoomData(ctx, rdb, true, nil)
			return confidencialRoomData.PublishUnsafeRoomData(ctx, rdb, false, confidencialPlayerIds)
		} else {
			r.PublishRoomUpdate(ctx, rdb, true, nil)
			return confidencialRoomData.PublishRoomUpdate(ctx, rdb, false, confidencialPlayerIds)
		}
	}
}

// Publica a versão privada completa, não segura da sala para todos os players de uma sala (APENAS PARA DEBUG)
func (r *Room) PublishUnsafeRoomData(ctx context.Context, rdb *redis.Client, toAll bool, playerIds []string) error {
	var broadcastMsg *structs.BroadcastMessage

	if toAll {
		broadcastMsg = structs.NewBroadcastMessage(r)
	} else {
		broadcastMsg = structs.NewSelectiveBroadcastMessage(r, playerIds)
	}

	if pubMsg, err := broadcastMsg.ToJSON(); err == nil {
		return rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// Publica a versão pública da sala para todos os players de uma sala
func (r *Room) PublishRoomUpdate(ctx context.Context, rdb *redis.Client, toAll bool, playerIds []string) error {
	// Prepara os dados públicos da sala
	roomDataPublic := PublicRoomForUpdates{
		ID:                r.ID,
		Turn:              r.Turn,
		Tax:               r.Tax,
		Players:           make(map[string]players.PublicPlayerForUpdates, len(r.Players)),
		Deck:              r.Deck,
		CurrentPlayer:     r.CurrentPlayer,
		GameEvent:         r.GameEvent,
		PendingEffects:    r.PendingEffects,
		GameOver:          r.GameOver,
		StartTime:         r.StartTime,
		DoubledCardValues: r.DoubledCardValues,
	}

	// Prepara os dados públicos dos jogadores
	for playerId, player := range r.Players {
		publicPlayer := player.GetPublicPlayerForUpdates()
		roomDataPublic.Players[playerId] = publicPlayer
	}

	// Cria a mensagem de broadcast
	var broadcastMsg *structs.BroadcastMessage
	if toAll {
		broadcastMsg = structs.NewBroadcastMessage(roomDataPublic)
	} else {
		broadcastMsg = structs.NewSelectiveBroadcastMessage(roomDataPublic, playerIds)
	}

	// Publica a mensagem
	if pubMsg, err := broadcastMsg.ToJSON(); err == nil {
		return rdb.Publish(ctx, "room:"+r.ID+":broadcast", pubMsg).Err()
	}
	return nil
}

// SetTTL define um TTL para uma sala específica
func (r *Room) SetTTL(ctx context.Context, rdb *redis.Client, ttl time.Duration) error {
	return rdb.Expire(ctx, "room:"+r.ID, ttl).Err()
}

// RemoveTTL remove o TTL de uma sala (torna ela persistente)
func (r *Room) RemoveTTL(ctx context.Context, rdb *redis.Client) error {
	return rdb.Persist(ctx, "room:"+r.ID).Err()
}

// CheckIfItsEmpty verifica se uma sala está vazia (sem jogadores conectados)
func (r *Room) CheckIfItsEmpty(ctx context.Context, rdb *redis.Client) (bool, error) {
	// Verifica se há jogadores conectados (registrados nesta sala)
	for playerId := range r.Players {
		registeredRoom, registered, err := playerRedis.GetRegisteredRoomForPlayer(ctx, rdb, playerId)
		if err != nil {
			return false, err
		}
		if registered && registeredRoom == r.ID {
			return false, nil
		}
	}

	return true, nil
}

// Assina o canal de broadcast de uma sala
func (r *Room) SubscribeRoomBroadcast(rdb *redis.Client) {
	if structs.ConnManager.IsSubscribed(r.ID) {
		utils.LogDebug("Sala " + r.ID + " já está assinada no Pub/Sub")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	structs.ConnManager.SetRoomCancel(r.ID, cancel)
	structs.ConnManager.MarkSubscribed(r.ID)

	pubsub := rdb.Subscribe(ctx, "room:"+r.ID+":broadcast")

	go func() {
		defer func() {
			_ = pubsub.Close()
			structs.ConnManager.ClearSubscribed(r.ID)
			utils.LogDebug("Pub/Sub encerrado da sala " + r.ID)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-pubsub.Channel():
				if !ok {
					return
				}

				// Tenta parsear como BroadcastMessage
				broadcastMsg, err := structs.ParseBroadcastMessage([]byte(msg.Payload))
				if err != nil {
					// Se falhar, envia a mensagem original para manter compatibilidade
					utils.LogError(fmt.Errorf("erro ao parsear BroadcastMessage, enviando payload original: %w", err))
					structs.ConnManager.Broadcast(r.ID, []byte(msg.Payload))
					continue
				}

				// Serializa apenas os dados para enviar aos clientes
				dataPayload, err := json.Marshal(broadcastMsg.Data)
				if err != nil {
					utils.LogError(fmt.Errorf("erro ao serializar dados da BroadcastMessage: %w", err))
					continue
				}

				// Decide se é broadcast para todos ou seletivo
				if broadcastMsg.ToAll {
					// Envia para todos os jogadores conectados na sala
					structs.ConnManager.Broadcast(r.ID, dataPayload)
				} else {
					// Envia apenas para os jogadores especificados
					structs.ConnManager.BroadcastSelective(r.ID, dataPayload, broadcastMsg.ConfidencialPlayerIds)
				}
			}
		}
	}()

	utils.LogDebug("Assinada sala " + r.ID + " no Pub/Sub")
}

func (r *Room) AppendPendingEffect(effect structs.Effect) error {
	dto, err := effect.ToDTO()
	if err != nil {
		return err
	}

	r.PendingEffects = append(r.PendingEffects, dto)
	return nil
}

func (r *Room) PopLastPendingEffect() (structs.EffectDTO, bool) {
	if len(r.PendingEffects) == 0 {
		return structs.EffectDTO{}, false
	}

	lastIdx := len(r.PendingEffects) - 1
	effect := r.PendingEffects[lastIdx]
	r.PendingEffects = r.PendingEffects[:lastIdx]
	return effect, true
}

// Verifica se jogador tem moedas suficientes para jogar a carta
func (r *Room) VerifyPlayerHasEnoughCoins(player *players.Player, registryRules *rules.Registry, card structs.TypePlayerPlays) error {
	cardPrice, err := r.GetCardValue(registryRules, card)
	if err != nil {
		return err
	}

	if player.Coins < cardPrice {
		return fmt.Errorf("jogador %s não tem moedas suficientes para jogar a carta %s (tem %d, precisa de %d)", player.Id, card, player.Coins, cardPrice)
	}

	return nil
}

// Marca como dobrado o preço da carta específica
func (r *Room) MarkCardValueAsDoubled(registryRules *rules.Registry, card structs.TypePlayerPlays) {
	timesDoubledCardConf, exist := r.DoubledCardValues[card]
	if !exist {
		return
	}

	cardRules, err := r.GetCardRules(registryRules, string(card))
	if err != nil {
		utils.LogError(err)
		return
	}

	if timesDoubledCardConf.TimesValueDoubled < *cardRules.MaxDoubled {
		// ainda pode dobrar o preço
		timesDoubledCardConf.TimesValueDoubled++
		timesDoubledCardConf.RoundsUntilDecrease = *cardRules.RoundsUntilDecrease
	} else {
		// já atingiu o máximo de vezes que pode ser dobrado, apenas reseta o contador de rounds
		timesDoubledCardConf.RoundsUntilDecrease = *cardRules.RoundsUntilDecrease
	}

	timesDoubledCardConf.UsedThisTurn = true

	r.DoubledCardValues[card] = timesDoubledCardConf
}

// DecreaseDoubledCardValuesRounds decrementa o contador de rounds para redução do preço dobrado das cartas e abaixa o preço se o contador estiver em 0
func (r *Room) DecreaseDoubledCardValuesRounds() {
	for cardType, timesDoubledCardConf := range r.DoubledCardValues {
		// decrementa o contador de rounds
		if !timesDoubledCardConf.UsedThisTurn {
			if timesDoubledCardConf.RoundsUntilDecrease > 0 {
				timesDoubledCardConf.RoundsUntilDecrease--
				r.DoubledCardValues[cardType] = timesDoubledCardConf
			} else {
				// diminui o preço dobrado em 1, se possível
				if timesDoubledCardConf.TimesValueDoubled > 0 {
					timesDoubledCardConf.TimesValueDoubled--
					r.DoubledCardValues[cardType] = timesDoubledCardConf
				}
			}
		} else {
			// reseta o flag de uso nesta rodada
			timesDoubledCardConf.UsedThisTurn = false
			r.DoubledCardValues[cardType] = timesDoubledCardConf
		}
	}
}

// GetCardValue retorna o value atual da carta, considerando se está dobrado ou não, taxas e limites.
//
// Taxas são aplicadas depois dos valores dobrados (ambos apenas se aplicáveis)
func (r *Room) GetCardValue(registryRules *rules.Registry, card structs.TypePlayerPlays) (int, error) {
	cardRules, err := r.GetCardRules(registryRules, string(card))
	if err != nil {
		return 0, err
	}

	value := *cardRules.Value

	// se a carta funcionar com preço dobrado, aplica o valor dobrado
	timesDoubledCardConf, exist := r.DoubledCardValues[card]
	if exist {
		value = value << timesDoubledCardConf.TimesValueDoubled
	}

	// se a carta for afetada por taxas, aplica a taxa
	if cardRules.AffectedByTaxes != nil && *cardRules.AffectedByTaxes {
		if value+r.Tax > *cardRules.TaxMaximum {
			// se a taxa ultrapassar os limites, aplica os limites
			value = *cardRules.TaxMaximum
		} else if value+r.Tax < *cardRules.TaxMinimum {
			// se a taxa for menor que o mínimo, aplica o mínimo
			value = *cardRules.TaxMinimum
		} else {
			// caso contrário, aplica a taxa normalmente
			value += r.Tax
		}
	}

	return value, nil
}

// Incrementa as taxas pelo valor passado
func (r *Room) IncrementTax(amount int) {
	r.Tax += amount
}

// Decrementa as taxas pelo valor passado
func (r *Room) DecrementTax(amount int) {
	r.Tax -= amount
}

func (r *Room) Clone() *Room {
	var clone Room
	b, _ := json.Marshal(r)
	_ = json.Unmarshal(b, &clone)
	return &clone
}
