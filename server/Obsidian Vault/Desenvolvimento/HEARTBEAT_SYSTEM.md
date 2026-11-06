# Sistema de Heartbeat com TTL Sincronizado

## 🔑 Como Funciona

O sistema implementado funciona da seguinte forma:

### 1. **Heartbeat da Instância**
- Cada instância envia heartbeat a cada 5 segundos
- O heartbeat é armazenado como: `instance:{instanceId}:alive` com TTL de 10 segundos
- Se a instância morrer, o heartbeat expira automaticamente

### 2. **Registro de Players**
- Quando um player se conecta, é registrado como: `player:{playerId}:room` → `{roomId}:{instanceId}`
- O registro tem TTL inicial de 10 segundos (mesmo do heartbeat)
- A cada heartbeat, a instância renova o TTL de todos os seus players

### 3. **Limpeza Automática**
- Se a instância morrer, o heartbeat expira
- Como não há mais renovação de TTL, os registros dos players também expiram
- O Redis limpa automaticamente sem necessidade de watchdog

## 🚀 Exemplo de Uso

### Verificar Status das Instâncias
```bash
curl http://localhost:8080/instances/status
```

Resposta:
```json
{
  "message": "Status das instâncias",
  "data": {
    "currentInstance": "abc123-def456",
    "totalInstances": 2,
    "instances": [
      {
        "instanceId": "abc123-def456",
        "isAlive": true,
        "managedPlayers": ["player1", "player2"]
      },
      {
        "instanceId": "xyz789-uvw012",
        "isAlive": true,
        "managedPlayers": ["player3"]
      }
    ],
    "orphanedPlayers": []
  }
}
```

### Limpar Players Órfãos (Manual)
```bash
curl -X POST http://localhost:8080/instances/cleanup
```

## 🔧 Configurações

As constantes podem ser ajustadas em `services/heartbeatService.go`:

```go
const (
    HeartbeatInterval = 5 * time.Second  // Intervalo entre heartbeats
    HeartbeatTTL      = 10 * time.Second // TTL do heartbeat e players
)
```

## 📊 Monitoramento

### Chaves Redis Criadas:
- `instance:{instanceId}:alive` - Heartbeat da instância
- `player:{playerId}:room` - Registro do player (formato: `roomId:instanceId`)

### Logs Gerados:
- `[Instance abc123] Heartbeat iniciado`
- `Heartbeat enviado - Instância: abc123, Players: 2`
- `[Instance abc123] Heartbeat parado`

## 🔄 Fluxo Completo

1. **Instância inicia** → HeartbeatService.Start()
2. **Player conecta** → RegisterPlayerInRoom() com TTL de 10s
3. **A cada 5s** → Heartbeat renova TTL próprio e dos players
4. **Instância morre** → Heartbeat para de renovar
5. **Após 10s** → Redis remove heartbeat e registros de players automaticamente

## 🛠️ Funcões Principais

### Services
- `HeartbeatService.Start()` - Inicia heartbeat
- `HeartbeatService.Stop()` - Para heartbeat
- `IsInstanceAlive()` - Verifica se instância está viva
- `GetAliveInstances()` - Lista instâncias vivas

### Redis Handlers
- `RegisterPlayerInRoom()` - Registra player com instanceId
- `GetRegisteredRoomForPlayer()` - Verifica registro e valida instância
- `GetPlayerRegistrationInfo()` - Informações completas do registro

### Utilitários
- `CleanupPlayerRegistrations()` - Limpeza manual de órfãos
- `LogInstanceInfo()` - Logs com identificação da instância
