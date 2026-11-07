package redisFuncs

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func newRoomLockRedis(t *testing.T) (*miniredis.Miniredis, *redis.Client) {
	t.Helper()
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = client.Close() })
	return mr, client
}

func TestAcquireRoomLockRenovaAlemDoTTL(t *testing.T) {
	mr, client := newRoomLockRedis(t)
	ctx := context.Background()
	const roomID = "sala-renova"
	ttl := 120 * time.Millisecond

	lock, err := AcquireRoomLock(ctx, client, roomID, "pod-a", ttl)
	if err != nil {
		t.Fatalf("adquirir lock: %v", err)
	}
	if lock == nil {
		t.Fatal("lock deveria ter sido adquirido")
	}
	defer lock.Release()

	// Deixa a renovação rodar algumas vezes em tempo real antes de avançar o relógio.
	time.Sleep(2 * (ttl / 3))

	// Avança o relógio do miniredis além do TTL em passos, dando chance à renovação
	// de reempurrar o PEXPIRE a cada passo. No total o relógio anda 6*(ttl/2) > ttl.
	for i := 0; i < 6; i++ {
		mr.FastForward(ttl / 2)
		time.Sleep(ttl / 3)
	}

	exists, err := client.Exists(ctx, "lock:room:"+roomID).Result()
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if exists != 1 {
		t.Fatalf("renovação não manteve o lock além do TTL (exists=%d)", exists)
	}
}

func TestAcquireRoomLockOutroDonoNaoConsegue(t *testing.T) {
	_, client := newRoomLockRedis(t)
	ctx := context.Background()
	const roomID = "sala-outro-dono"
	ttl := time.Second

	lock, err := AcquireRoomLock(ctx, client, roomID, "pod-a", ttl)
	if err != nil {
		t.Fatalf("lock pod-a: %v", err)
	}
	if lock == nil {
		t.Fatal("lock pod-a deveria ter sido adquirido")
	}
	defer lock.Release()

	other, err := AcquireRoomLock(ctx, client, roomID, "pod-b", ttl)
	if err != nil {
		t.Fatalf("lock pod-b: %v", err)
	}
	if other != nil {
		other.Release()
		t.Fatal("pod-b não deveria conseguir o lock enquanto pod-a renova")
	}
}

func TestReleaseRoomLockNaoDonoNaoApaga(t *testing.T) {
	_, client := newRoomLockRedis(t)
	ctx := context.Background()
	const roomID = "sala-release"
	key := "lock:room:" + roomID
	ttl := time.Second

	lock, err := AcquireRoomLock(ctx, client, roomID, "pod-a", ttl)
	if err != nil {
		t.Fatalf("lock pod-a: %v", err)
	}
	if lock == nil {
		t.Fatal("lock pod-a deveria ter sido adquirido")
	}
	defer lock.Release()

	// Outro dono tenta soltar via Lua compare-and-set: não pode apagar.
	if _, err := releaseRoomLockScript.Run(ctx, client, []string{key}, "pod-b").Result(); err != nil {
		t.Fatalf("release não-dono: %v", err)
	}

	exists, err := client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("exists: %v", err)
	}
	if exists != 1 {
		t.Fatal("release de não-dono apagou o lock de outro pod")
	}
}

func TestPerdaDoLockCancelaCtx(t *testing.T) {
	_, client := newRoomLockRedis(t)
	ctx := context.Background()
	const roomID = "sala-perda"
	ttl := time.Second

	lock, err := AcquireRoomLock(ctx, client, roomID, "pod-a", ttl)
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	if lock == nil {
		t.Fatal("lock deveria ter sido adquirido")
	}
	defer lock.Release()

	// Apaga a chave: a próxima renovação perde o lock e cancela o ctx derivado.
	if err := client.Del(ctx, "lock:room:"+roomID).Err(); err != nil {
		t.Fatalf("del: %v", err)
	}

	deadline := time.Now().Add(2 * time.Second)
	for lock.Err() == nil && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if lock.Err() == nil {
		t.Fatal("ctx não foi cancelado após a perda do lock")
	}
}

func TestPodMortoLiberaEmTTL(t *testing.T) {
	mr, client := newRoomLockRedis(t)
	const roomID = "sala-pod-morto"
	ttl := 120 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	lock, err := AcquireRoomLock(ctx, client, roomID, "pod-a", ttl)
	if err != nil {
		t.Fatalf("lock pod-a: %v", err)
	}
	if lock == nil {
		t.Fatal("lock pod-a deveria ter sido adquirido")
	}

	// Pod "morre": cancela o ctx, parando a renovação sem soltar o lock.
	cancel()
	<-lock.done

	// O lock vence em ~TTL e outro pod assume.
	mr.FastForward(ttl + time.Millisecond)

	other, err := AcquireRoomLock(context.Background(), client, roomID, "pod-b", ttl)
	if err != nil {
		t.Fatalf("lock pod-b: %v", err)
	}
	if other == nil {
		t.Fatal("pod morto não liberou a sala após o TTL")
	}
	other.Release()
}

func TestRenewInterval(t *testing.T) {
	casos := []struct {
		ttl  time.Duration
		want time.Duration
	}{
		{9 * time.Second, 3 * time.Second},              // ttl/3 > 1s
		{3 * time.Second, time.Second},                  // max(ttl/3, 1s)
		{2 * time.Second, time.Second},                  // max(ttl/3, 1s)
		{120 * time.Millisecond, 40 * time.Millisecond}, // TTL pequeno: ttl/3, senão venceria
	}
	for _, caso := range casos {
		if got := renewInterval(caso.ttl); got != caso.want {
			t.Fatalf("renewInterval(%v) = %v, quer %v", caso.ttl, got, caso.want)
		}
	}
}

func TestLockCurtoNaoRenovaAntesDoRelease(t *testing.T) {
	_, client := newRoomLockRedis(t)
	ctx := context.Background()

	counter := newCommandCounter()
	client.AddHook(counter)

	lock, err := AcquireRoomLock(ctx, client, "sala-curta", "pod-a", 3*time.Second)
	if err != nil {
		t.Fatalf("lock: %v", err)
	}
	if lock == nil {
		t.Fatal("lock deveria ter sido adquirido")
	}

	// fn curto: o timer de 1s nem dispara, então nenhuma renovação chega ao Redis.
	time.Sleep(50 * time.Millisecond)
	if got := counter.get("eval") + counter.get("evalsha"); got != 0 {
		t.Fatalf("houve %d renovação(ões) antes do release", got)
	}

	if err := lock.Release(); err != nil {
		t.Fatalf("release: %v", err)
	}
	if got := counter.get("eval") + counter.get("evalsha"); got == 0 {
		t.Fatal("release deveria rodar o Lua compare-and-set")
	}
}
