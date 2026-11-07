package redisFuncs

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

const roomLockKeyPrefix = "lock:room:"

// Lua compare-and-set do dono: só renova/solta se a chave ainda for deste dono.
// Fecha a corrida do antigo Get+Del separado, que podia apagar o lock de outro pod.
var (
	renewRoomLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("PEXPIRE", KEYS[1], ARGV[2])
end
return 0`)

	releaseRoomLockScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0`)
)

// RoomLock mantém o lock distribuído de uma sala e o renova em segundo plano.
type RoomLock struct {
	rdb      *redis.Client
	key      string
	owner    string
	ttl      time.Duration
	interval time.Duration

	cancel context.CancelFunc
	done   chan struct{}
	ctx    context.Context
	timer  *time.Timer
}

// AcquireRoomLock tenta adquirir o lock distribuído da sala. Se a sala já pertence
// a outro dono, retorna (nil, nil). O lock é renovado em segundo plano enquanto o
// ctx não for cancelado; se a renovação falhar (lock perdido), o ctx derivado é
// cancelado e a goroutine de renovação encerra.
func AcquireRoomLock(ctx context.Context, rdb *redis.Client, roomID, instanceID string, ttl time.Duration) (*RoomLock, error) {
	key := roomLockKeyPrefix + roomID

	ok, err := rdb.SetNX(ctx, key, instanceID, ttl).Result()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	renewCtx, cancel := context.WithCancel(ctx)
	interval := renewInterval(ttl)

	lock := &RoomLock{
		rdb:      rdb,
		key:      key,
		owner:    instanceID,
		ttl:      ttl,
		interval: interval,
		cancel:   cancel,
		done:     make(chan struct{}),
		ctx:      renewCtx,
	}
	// Só cria goroutine/ticker quando a primeira renovação for realmente necessária:
	// fn curto (a maioria) termina antes do timer e não paga a renovação.
	lock.timer = time.AfterFunc(interval, func() { lock.renew(renewCtx) })
	return lock, nil
}

// renewInterval devolve o intervalo da primeira renovação (e dos ticks seguintes):
// max(ttl/3, 1s), mas nunca a partir do TTL — com TTL pequeno (testes), renovar em
// 1s deixaria o lock vencer antes, então cai para ttl/3.
func renewInterval(ttl time.Duration) time.Duration {
	interval := ttl / 3
	if interval < time.Second {
		interval = time.Second
	}
	if interval >= ttl {
		interval = ttl / 3
	}
	if interval <= 0 {
		interval = time.Millisecond
	}
	return interval
}

func (l *RoomLock) renew(ctx context.Context) {
	defer close(l.done)

	if ctx.Err() != nil {
		return
	}
	if !l.renewOnce(ctx) {
		return
	}

	ticker := time.NewTicker(l.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if !l.renewOnce(ctx) {
				return
			}
		}
	}
}

// renewOnce estende o lock; devolve false quando ele foi perdido (e cancela o ctx).
func (l *RoomLock) renewOnce(ctx context.Context) bool {
	ok, err := renewRoomLockScript.Run(ctx, l.rdb, []string{l.key}, l.owner, l.ttl.Milliseconds()).Int()
	if err != nil || ok == 0 {
		// Lock perdido: cancela o ctx para que SaveRoom não grave por cima de outro pod.
		l.cancel()
		return false
	}
	return true
}

// Err retorna o erro do ctx derivado (context.Canceled) quando o lock foi perdido.
func (l *RoomLock) Err() error {
	return l.ctx.Err()
}

// Context retorna o contexto derivado usado pela renovação em segundo plano.
func (l *RoomLock) Context() context.Context {
	return l.ctx
}

// Release para a renovação e solta o lock se ainda pertencer a este dono.
func (l *RoomLock) Release() error {
	l.cancel()

	// Se o timer ainda não disparou, nenhuma goroutine de renovação chegou a existir.
	if l.timer != nil && l.timer.Stop() {
		close(l.done)
	}
	<-l.done

	// Solta com um ctx próprio: o defer do WithRoom precisa funcionar mesmo que o ctx
	// do chamador já tenha sido cancelado (se falhar, o lock expira sozinho pelo TTL).
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := releaseRoomLockScript.Run(ctx, l.rdb, []string{l.key}, l.owner).Result()
	return err
}
