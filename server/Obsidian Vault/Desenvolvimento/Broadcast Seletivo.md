# Broadcast Seletivo em Hoodwink

## Visão Geral

O sistema de broadcast foi modificado para suportar envio de atualizações de sala tanto para todos os jogadores quanto para jogadores específicos. Isso permite enviar informações confidenciais apenas para os jogadores que devem vê-las.

## Estrutura BroadcastMessage

A estrutura `BroadcastMessage` encapsula mensagens com metadados de roteamento:

```go
type BroadcastMessage struct {
    ToAll                 bool     `json:"toAll"`
    ConfidencialPlayerIds []string `json:"confidencialPlayerIds,omitempty"`
    Data                  any      `json:"data"`
}
```

## Como Usar

### Broadcast para Todos (Comportamento Padrão)

```go
// Envia atualização pública para todos os jogadores
err := roomData.SendUpdatedRoomData(ctx, rdb, nil, []string{})
```

### Broadcast Seletivo

```go
// Cria dados confidenciais da sala
confidencialRoomData := &rooms.Room{
    // Preenche com dados confidenciais específicos
    // (ex: cartas de um jogador específico)
}

// Lista de IDs dos jogadores que devem receber os dados confidenciais
playerIds := []string{"player1", "player2"}

// Envia atualização confidencial apenas para os jogadores especificados
err := roomData.SendUpdatedRoomData(ctx, rdb, confidencialRoomData, playerIds)
```

## Fluxo de Funcionamento

1. **Publicação**: `SendUpdatedRoomData` cria uma `BroadcastMessage` com:
   - `ToAll: true` se `confidencialRoomData` for `nil` OU `confidencialPlayerIds` estiver vazio
   - `ToAll: false` caso contrário, incluindo a lista de `confidencialPlayerIds`

2. **Redis Pub/Sub**: A mensagem é publicada no canal Redis `room:{roomId}:broadcast`

3. **Listener**: O `SubscribeRoomBroadcast` processa a mensagem:
   - Se `ToAll: true`, chama `ConnManager.Broadcast()` (envia para todos)
   - Se `ToAll: false`, chama `ConnManager.BroadcastSelective()` (envia apenas para os playerIds listados)

4. **Entrega**: Apenas as conexões WebSocket dos jogadores especificados recebem a mensagem

## Exemplo de Uso Real

### Clairvoyant Card (Futuro)

Quando a carta Clairvoyant for implementada, o jogador que a usar poderá ver as cartas de outro jogador:

```go
func ClairvoyantEffect(ctx context.Context, rdb *redis.Client, registryRules *rules.Registry, roomData *rooms.Room, effect structs.Effect) *PrivateUpdate {
    // ... lógica do efeito ...
    
    // Cria sala confidencial com informações adicionais
    confidencialRoomData := *roomData
    targetPlayer, _ := confidencialRoomData.GetPlayer(targetPlayerId)
    
    // Apenas o sourcePlayer vê as cartas do targetPlayer
    return &PrivateUpdate{
        ConfidencialRoomData: &confidencialRoomData,
        PlayerIds:            []string{sourcePlayerId},
    }
}

// No resolveEffects.go
if privateUpdate != nil {
    roomData.SendUpdatedRoomData(ctx, rdb, privateUpdate.ConfidencialRoomData, privateUpdate.PlayerIds)
}
```

## Métodos do ConnectionManager

### `Broadcast(roomId string, message []byte)`
Envia mensagem para todos os jogadores conectados na sala.

### `BroadcastSelective(roomId string, message []byte, playerIds []string)`
Envia mensagem apenas para os jogadores especificados que estão conectados na sala.

## Observações Importantes

1. **Retrocompatibilidade**: Se o parsing de `BroadcastMessage` falhar, o sistema envia a mensagem original para manter compatibilidade.

2. **Performance**: O `BroadcastSelective` usa um map para busca eficiente dos playerIds alvo.

3. **Segurança**: Os dados confidenciais nunca são enviados para jogadores não autorizados.

4. **Debug Mode**: No modo debug unsafe, a sala completa (com todos os dados) pode ser enviada, respeitando o broadcast seletivo se configurado.

## Arquivos Modificados

- `gameServer/game/structs/broadcast.go` (novo)
- `gameServer/game/structs/connManager.go`
- `gameServer/game/structs/rooms/room.go`
- Todas as chamadas de `SendUpdatedRoomData` no projeto

## Próximos Passos

Para implementar uma nova carta com informações confidenciais:

1. Crie o effect correspondente retornando `*PrivateUpdate` (se necessário criar essa struct)
2. No `resolveEffects.go`, use o retorno para chamar `SendUpdatedRoomData` com os dados confidenciais
3. No frontend, trate a mensagem normalmente - ela já chegará filtrada para os jogadores corretos
