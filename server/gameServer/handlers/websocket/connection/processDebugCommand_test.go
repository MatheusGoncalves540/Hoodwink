package connection

import (
	"context"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs"
	"github.com/MatheusGoncalves540/Hoodwink-gameServer/game/structs/rooms"
)

// TestProcessDebugCommandDisabledByDefault garante que, sem UNSAFE_DEBUG, um
// comando de debug não altera o estado da sala e não toca o Redis (o rdb nil
// estouraria caso qualquer comando tentasse salvar/transmitir).
func TestProcessDebugCommandDisabledByDefault(t *testing.T) {
	t.Setenv("UNSAFE_DEBUG", "")

	room := &rooms.Room{
		ID:        "r1",
		Turn:      1,
		StartTime: time.Time{},
	}
	play := &structs.PlayerPlay{Type: "COMMAND_PRINT_ROOM", PlayerId: "p1"}

	ProcessDebugCommand(context.Background(), nil, room, play)

	if !room.StartTime.IsZero() {
		t.Fatalf("sala não deveria ter mudado com UNSAFE_DEBUG desligado, StartTime=%v", room.StartTime)
	}
	if room.Turn != 1 {
		t.Fatalf("sala não deveria ter mudado com UNSAFE_DEBUG desligado, Turn=%d", room.Turn)
	}
}
