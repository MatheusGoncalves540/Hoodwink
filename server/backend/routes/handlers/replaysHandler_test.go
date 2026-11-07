package handlers

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink-replay/format"
	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// --- fakes ---

type fakeReplayStore struct {
	inserted  map[string]*models.MatchReplay // chave = MatchID; primeiro vence (DO NOTHING)
	insertErr error
	listFn    func(ctx context.Context, playerID string, limit int, since time.Time) ([]models.MatchReplay, error)
	findFn    func(ctx context.Context, matchID, playerID string) (*models.MatchReplay, error)
}

func newFakeReplayStore() *fakeReplayStore {
	return &fakeReplayStore{inserted: map[string]*models.MatchReplay{}}
}

func (f *fakeReplayStore) Insert(ctx context.Context, r *models.MatchReplay) error {
	if f.insertErr != nil {
		return f.insertErr
	}
	if _, exists := f.inserted[r.MatchID]; !exists {
		cp := *r
		cp.PlayerIDs = append(models.StringArray(nil), r.PlayerIDs...)
		cp.Data = append([]byte(nil), r.Data...)
		if r.WinnerID != nil {
			w := *r.WinnerID
			cp.WinnerID = &w
		}
		f.inserted[r.MatchID] = &cp
	}
	return nil
}

func (f *fakeReplayStore) ListByPlayer(ctx context.Context, playerID string, limit int, since time.Time) ([]models.MatchReplay, error) {
	if f.listFn != nil {
		return f.listFn(ctx, playerID, limit, since)
	}
	return nil, nil
}

func (f *fakeReplayStore) FindByMatchAndPlayer(ctx context.Context, matchID, playerID string) (*models.MatchReplay, error) {
	if f.findFn != nil {
		return f.findFn(ctx, matchID, playerID)
	}
	return nil, gorm.ErrRecordNotFound
}

// --- helpers ---

var testMatchUUID = uuid.MustParse("0193a1c0-0000-7000-8000-000000000001")

var (
	testPlayerA = uuid.MustParse("11111111-1111-1111-1111-111111111111").String()
	testPlayerB = uuid.MustParse("22222222-2222-2222-2222-222222222222").String()
	testPlayerC = uuid.MustParse("33333333-3333-3333-3333-333333333333").String()
	testPlayerD = uuid.MustParse("44444444-4444-4444-4444-444444444444").String()
)

func testSigner(t *testing.T) (format.Signer, ed25519.PublicKey) {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("generate ed25519: %v", err)
	}
	return format.NewEd25519Signer("replay-v1", priv), pub
}

func buildReplay(t *testing.T, winner int, withEnd bool) []byte {
	t.Helper()
	players := []format.Player{
		{Index: 0, Name: "ana"},
		{Index: 1, Name: "bia"},
	}
	entries := []format.Entry{
		{DeltaMs: 0, Type: format.EntryStart, Player: format.NoPlayer},
		{DeltaMs: 1000, Type: format.EntryPlay, Player: 0, Payload: []byte(`{"type":"TAKE_COINS"}`)},
	}
	if withEnd {
		entries = append(entries, format.Entry{DeltaMs: 1000, Type: format.EntryEnd, Player: uint8(winner)})
	}
	r := &format.Replay{
		EngineVersion: 1,
		MatchID:       [16]byte(testMatchUUID),
		Seed:          [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "classic",
		Players:       players,
		Entries:       entries,
	}
	n := (len(entries) + format.CheckpointInterval - 1) / format.CheckpointInterval
	r.Checkpoints = make([]format.Checkpoint, n)
	for i := range r.Checkpoints {
		r.Checkpoints[i] = format.Checkpoint{byte(i), 1, 2, 3, 4, 5, 6, 7}
	}
	data, err := format.Encode(r, nil)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	return data
}

func postUpload(store replayStore, signer format.Signer, body map[string]any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	_ = json.NewEncoder(&buf).Encode(body)
	req := httptest.NewRequest(http.MethodPost, "/internal/replays", &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	writeReplayUpload(rr, req, store, signer)
	return rr
}

func uploadBody(raw []byte, playerIDs []string, roomID string) map[string]any {
	return map[string]any{
		"data":       base64.StdEncoding.EncodeToString(raw),
		"room_id":    roomID,
		"player_ids": playerIDs,
	}
}

func replayGET(matchID, userID string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("matchId", matchID)
	req := httptest.NewRequest(http.MethodGet, "/replays/"+matchID, nil)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	if userID != "" {
		req = req.WithContext(context.WithValue(req.Context(), UserContextKey, structures.UserClaims{Id: userID}))
	}
	return req
}

// --- testes ---

func TestReplayUploadSemFim409(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	raw := buildReplay(t, 1, false) // sem EntryEnd
	rr := postUpload(store, signer, uploadBody(raw, []string{testPlayerA, testPlayerB}, "r1"))

	if rr.Code != http.StatusConflict {
		t.Fatalf("esperado 409 para replay sem fim, veio %d: %s", rr.Code, rr.Body.String())
	}
	if len(store.inserted) != 0 {
		t.Fatalf("não deveria gravar replay sem fim, gravou %d", len(store.inserted))
	}
}

func TestReplayUploadAcimaDe300KB413(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	body := map[string]any{
		"data":       strings.Repeat("A", 300*1024),
		"room_id":    "r1",
		"player_ids": []string{testPlayerA, testPlayerB},
	}
	rr := postUpload(store, signer, body)

	if rr.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("esperado 413 para 300 KB, veio %d", rr.Code)
	}
	if len(store.inserted) != 0 {
		t.Fatalf("não deveria gravar acima de 256 KB, gravou %d", len(store.inserted))
	}
}

func TestReplayUploadMesmoMatchIdIdempotente(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	raw := buildReplay(t, 1, true)
	rr1 := postUpload(store, signer, uploadBody(raw, []string{testPlayerA, testPlayerB}, "r1"))
	if rr1.Code != http.StatusOK {
		t.Fatalf("primeiro POST deveria ser 200, veio %d: %s", rr1.Code, rr1.Body.String())
	}

	// Mesmo matchId com player_ids diferentes: não pode sobrescrever.
	rr2 := postUpload(store, signer, uploadBody(raw, []string{testPlayerC, testPlayerD}, "r1"))
	if rr2.Code != http.StatusOK {
		t.Fatalf("segundo POST deveria ser 200, veio %d: %s", rr2.Code, rr2.Body.String())
	}

	if len(store.inserted) != 1 {
		t.Fatalf("esperado 1 replay gravado, veio %d", len(store.inserted))
	}
	got := store.inserted[testMatchUUID.String()]
	if got == nil {
		t.Fatal("replay não encontrado no store")
	}
	if got.PlayerIDs[0] != testPlayerA {
		t.Fatalf("player_ids sobrescrito: veio %v, esperado %q", got.PlayerIDs, testPlayerA)
	}
}

func TestReplayUploadPlayerIDsEVencedorDoCorpo(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	// Vencedor do cabeçalho é o índice 1 (bia); o corpo não manda winner.
	raw := buildReplay(t, 1, true)
	body := uploadBody(raw, []string{testPlayerA, testPlayerB}, "r1")
	body["winner"] = testPlayerA // ignorado: vencedor vem do cabeçalho

	rr := postUpload(store, signer, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}
	got := store.inserted[testMatchUUID.String()]
	if got.WinnerID == nil || *got.WinnerID != testPlayerB {
		t.Fatalf("winner_id = %v, esperado %q (vencedor do cabeçalho, não do corpo)", got.WinnerID, testPlayerB)
	}
}

func TestReplayUploadPlayerIDsContagemInvalida400(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	raw := buildReplay(t, 1, true)
	// 3 UUIDs para 2 jogadores no cabeçalho → 400.
	rr := postUpload(store, signer, uploadBody(raw, []string{testPlayerA, testPlayerB, testPlayerC}, "r1"))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400 para player_ids que não confere, veio %d", rr.Code)
	}
}

func TestReplayUploadVersaoDesconhecida400(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	// Forja uma versão de motor inexistente (99) re-encodando manualmente.
	r := &format.Replay{
		EngineVersion: 99,
		MatchID:       [16]byte(testMatchUUID),
		Seed:          [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		StartedAtMs:   1700000000000,
		Mode:          "classic",
		Players: []format.Player{
			{Index: 0, Name: "ana"},
			{Index: 1, Name: "bia"},
		},
		Entries: []format.Entry{
			{DeltaMs: 0, Type: format.EntryStart, Player: format.NoPlayer},
			{DeltaMs: 1000, Type: format.EntryEnd, Player: 1},
		},
		Checkpoints: []format.Checkpoint{{1, 2, 3, 4, 5, 6, 7, 8}},
	}
	v99, err := format.Encode(r, nil)
	if err != nil {
		t.Fatalf("encode v99: %v", err)
	}

	rr := postUpload(store, signer, uploadBody(v99, []string{testPlayerA, testPlayerB}, "r1"))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400 para versão desconhecida, veio %d", rr.Code)
	}
}

func TestReplayDownloadNaoParticipanteEInexistente404(t *testing.T) {
	store := newFakeReplayStore()
	// Nenhum replay cadastrado: inexistente e não participante dão o mesmo 404.
	store.findFn = func(ctx context.Context, matchID, playerID string) (*models.MatchReplay, error) {
		return nil, gorm.ErrRecordNotFound
	}

	for _, id := range []string{testMatchUUID.String(), uuid.NewString()} {
		rr := httptest.NewRecorder()
		writeReplay(rr, replayGET(id, testPlayerA), store)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("esperado 404 para %s, veio %d", id, rr.Code)
		}
	}
}

func TestReplayDownloadNoStore(t *testing.T) {
	store := newFakeReplayStore()
	w := testPlayerB
	record := &models.MatchReplay{
		MatchID:   testMatchUUID.String(),
		EndedAt:   time.Date(2026, 11, 6, 12, 0, 0, 0, time.UTC),
		WinnerID:  &w,
		PlayerIDs: models.StringArray{testPlayerA, testPlayerB},
		Data:      []byte{1, 2, 3},
	}
	store.findFn = func(ctx context.Context, matchID, playerID string) (*models.MatchReplay, error) {
		return record, nil
	}

	rr := httptest.NewRecorder()
	writeReplay(rr, replayGET(testMatchUUID.String(), testPlayerA), store)

	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d", rr.Code)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "private, no-store" {
		t.Fatalf("Cache-Control = %q, esperado private, no-store", cc)
	}
	if cd := rr.Header().Get("Content-Disposition"); !strings.Contains(cd, "hoodwink-2026-11-06-0193a1c0.hwr") {
		t.Fatalf("Content-Disposition = %q", cd)
	}
}

func TestMeMatchesIgnoraUserIdQuery(t *testing.T) {
	store := newFakeReplayStore()
	var gotPlayer string
	store.listFn = func(ctx context.Context, playerID string, limit int, since time.Time) ([]models.MatchReplay, error) {
		gotPlayer = playerID
		return nil, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/me/matches?userId=outro&limit=5", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserContextKey, structures.UserClaims{Id: testPlayerA}))
	rr := httptest.NewRecorder()
	writeMeMatches(rr, req, store)

	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}
	if gotPlayer != testPlayerA {
		t.Fatalf("usou userId da query (%q) em vez do token (%q)", gotPlayer, testPlayerA)
	}
}

func TestReplayUploadAssinaturaConfere(t *testing.T) {
	store := newFakeReplayStore()
	signer, pub := testSigner(t)

	raw := buildReplay(t, 1, true)
	rr := postUpload(store, signer, uploadBody(raw, []string{testPlayerA, testPlayerB}, "r1"))
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}

	got := store.inserted[testMatchUUID.String()]
	if got == nil {
		t.Fatal("replay não gravado")
	}

	decoded, err := format.Decode(got.Data, format.StaticKeys(map[string]ed25519.PublicKey{"replay-v1": pub}))
	if err != nil {
		t.Fatalf("assinatura não confere: %v", err)
	}
	if !decoded.Official {
		t.Fatal("assinatura deveria marcar Official=true")
	}
	if decoded.Kid != "replay-v1" {
		t.Fatalf("kid = %q, esperado replay-v1", decoded.Kid)
	}
}
