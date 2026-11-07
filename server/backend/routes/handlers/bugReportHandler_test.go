package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/MatheusGoncalves540/Hoodwink/services"
	"github.com/MatheusGoncalves540/Hoodwink/structures"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// v7UUID constrói um UUID v7 com o timestamp dado.
func v7UUID(t *testing.T, at time.Time) string {
	t.Helper()
	b := make([]byte, 16)
	ms := at.UnixMilli()
	b[0] = byte(ms >> 40)
	b[1] = byte(ms >> 32)
	b[2] = byte(ms >> 24)
	b[3] = byte(ms >> 16)
	b[4] = byte(ms >> 8)
	b[5] = byte(ms)
	b[6] = 0x70 // versão 7 no nibble alto
	b[8] = 0x80 // variante RFC 4122
	u, err := uuid.FromBytes(b)
	if err != nil {
		t.Fatalf("uuid from bytes: %v", err)
	}
	return u.String()
}

func bugJSONRequest(method, target string, body any) *http.Request {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, target, &buf)
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:1234"
	return req
}

func TestValidMatchID(t *testing.T) {
	now := time.Now().UTC()

	if !validMatchID(v7UUID(t, now), now) {
		t.Fatal("v7 de agora deveria ser válido")
	}
	if !validMatchID(v7UUID(t, now.Add(-23*time.Hour)), now) {
		t.Fatal("v7 de 23h atrás deveria ser válido")
	}
	if !validMatchID(v7UUID(t, now.Add(3*time.Minute)), now) {
		t.Fatal("v7 de 3min à frente deveria ser válido")
	}
	if validMatchID(v7UUID(t, now.Add(-25*time.Hour)), now) {
		t.Fatal("v7 de 25h atrás deveria ser inválido")
	}
	if validMatchID(v7UUID(t, now.Add(6*time.Minute)), now) {
		t.Fatal("v7 de 6min à frente deveria ser inválido")
	}
	if validMatchID(uuid.NewString(), now) {
		t.Fatal("uuid v4 deveria ser inválido")
	}
	if validMatchID("nao-e-uuid", now) {
		t.Fatal("não-UUID deveria ser inválido")
	}
}

func TestValidBugCategory(t *testing.T) {
	if !validBugCategory("moedas") || !validBugCategory("outro") {
		t.Fatal("categorias válidas deveriam passar")
	}
	if validBugCategory("") || validBugCategory("hack") {
		t.Fatal("categoria inválida deveria falhar")
	}
}

func TestSanitizeBugContext(t *testing.T) {
	got, err := sanitizeBugContext(json.RawMessage(`{"a":{"ticket":"abc","n":1}}`))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if !bytes.Contains(got, []byte(`"***"`)) || bytes.Contains(got, []byte(`abc`)) {
		t.Fatalf("ticket aninhado deveria virar ***, veio %s", got)
	}

	if got, err := sanitizeBugContext(json.RawMessage(`null`)); err != nil || got != nil {
		t.Fatalf("null deveria virar nil: got=%s err=%v", got, err)
	}
	if _, err := sanitizeBugContext(json.RawMessage(`{`)); err == nil {
		t.Fatal("JSON inválido deveria errar")
	}
	big := make([]byte, maxBugContextBytes+1)
	for i := range big {
		big[i] = 'x'
	}
	big = append([]byte(`"`), append(big, '"')...)
	if _, err := sanitizeBugContext(json.RawMessage(big)); err == nil {
		t.Fatal("contexto acima de 8 KB deveria errar")
	}
}

func TestDecodeBugBodyMaxBytes(t *testing.T) {
	w := httptest.NewRecorder()
	desc := strings.Repeat("a", 20*1024)
	body := `{"category":"outro","description":"` + desc + `"}`
	req := httptest.NewRequest(http.MethodPost, "/bugReports", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	var dst any
	if decodeBugBody(w, req, &dst) {
		t.Fatal("corpo de 20 KB deveria ser recusado")
	}
	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("esperado 413, veio %d", w.Code)
	}
}

// --- flag ---

type fakeFlagStore struct {
	count    int64
	countErr error
	upsertFn func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error)
}

func (f *fakeFlagStore) CountPartida24h(ctx context.Context, userID string) (int64, error) {
	return f.count, f.countErr
}

func (f *fakeFlagStore) UpsertTurn(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
	return f.upsertFn(ctx, userID, username, matchID, ip, ua, turn)
}

func doFlag(t *testing.T, store bugFlagStore, body map[string]any, now time.Time) *httptest.ResponseRecorder {
	t.Helper()
	req := bugJSONRequest(http.MethodPost, "/bugReports/flag", body)
	rr := httptest.NewRecorder()
	flagBugReport(rr, req, structures.UserClaims{Id: "u1", Username: "ana"}, store, now)
	return rr
}

func TestFlagSemMatchIDValido400(t *testing.T) {
	store := &fakeFlagStore{upsertFn: func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
		t.Fatal("upsert não deveria ser chamado")
		return 0, nil, false, nil
	}}
	rr := doFlag(t, store, map[string]any{"matchId": "nada", "turn": 3}, time.Now())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

func TestFlagTurnoInvalido400(t *testing.T) {
	store := &fakeFlagStore{upsertFn: func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
		t.Fatal("upsert não deveria ser chamado")
		return 0, nil, false, nil
	}}
	m := v7UUID(t, time.Now())
	rr := doFlag(t, store, map[string]any{"matchId": m, "turn": 10001}, time.Now())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

func TestFlag51Relato429(t *testing.T) {
	store := &fakeFlagStore{count: 50, upsertFn: func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
		t.Fatal("upsert não deveria ser chamado")
		return 0, nil, false, nil
	}}
	m := v7UUID(t, time.Now())
	rr := doFlag(t, store, map[string]any{"matchId": m, "turn": 3}, time.Now())
	if rr.Code != http.StatusTooManyRequests {
		t.Fatalf("esperado 429, veio %d", rr.Code)
	}
}

func TestFlagCria201(t *testing.T) {
	store := &fakeFlagStore{upsertFn: func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
		return 7, []int32{3}, true, nil
	}}
	m := v7UUID(t, time.Now())
	rr := doFlag(t, store, map[string]any{"matchId": m, "turn": 3}, time.Now())
	if rr.Code != http.StatusCreated {
		t.Fatalf("esperado 201, veio %d: %s", rr.Code, rr.Body.String())
	}
}

func TestFlagMesmoTurnoIdempotente200(t *testing.T) {
	store := &fakeFlagStore{upsertFn: func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
		return 7, []int32{1, 3}, false, nil
	}}
	m := v7UUID(t, time.Now())
	rr := doFlag(t, store, map[string]any{"matchId": m, "turn": 3}, time.Now())
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d", rr.Code)
	}
}

func TestFlag11Turno400(t *testing.T) {
	turns := make([]int32, 10)
	for i := range turns {
		turns[i] = int32(i)
	}
	store := &fakeFlagStore{upsertFn: func(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
		return 7, turns, false, nil // turno 11 não entra
	}}
	m := v7UUID(t, time.Now())
	rr := doFlag(t, store, map[string]any{"matchId": m, "turn": 11}, time.Now())
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

// --- lobby ---

type fakeCreateStore struct {
	last *models.BugReport
	err  error
}

func (f *fakeCreateStore) Create(ctx context.Context, report *models.BugReport) error {
	f.last = report
	if report.ID == 0 {
		report.ID = 1
	}
	return f.err
}

func doLobby(t *testing.T, store bugCreateStore, logged bool, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	req := bugJSONRequest(http.MethodPost, "/bugReports", body)
	rr := httptest.NewRecorder()
	createLobbyBug(rr, req, structures.UserClaims{Id: "u1", Username: "ana"}, logged, store)
	return rr
}

func TestLobbyCategoriaObrigatoria400(t *testing.T) {
	rr := doLobby(t, &fakeCreateStore{}, false, map[string]any{"category": "hack", "description": "descricao longa o suficiente"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

func TestLobbyDescricaoCurta400(t *testing.T) {
	rr := doLobby(t, &fakeCreateStore{}, false, map[string]any{"category": "outro", "description": "curto"})
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

func TestLobbyAnonimo201UserNulo(t *testing.T) {
	store := &fakeCreateStore{}
	rr := doLobby(t, store, false, map[string]any{"category": "outro", "description": "descricao suficiente", "context": map[string]any{"ticket": "segredo"}})
	if rr.Code != http.StatusCreated {
		t.Fatalf("esperado 201, veio %d: %s", rr.Code, rr.Body.String())
	}
	if store.last == nil {
		t.Fatal("relato não gravado")
	}
	if store.last.UserID != nil {
		t.Fatalf("anônimo deveria ter user_id nulo, veio %v", *store.last.UserID)
	}
	if store.last.Source != "lobby" {
		t.Fatalf("source = %q, esperado lobby", store.last.Source)
	}
	if bytes.Contains(store.last.Context, []byte("segredo")) {
		t.Fatalf("contexto não mascarado: %s", store.last.Context)
	}
}

func TestLobbyLogadoGravaUsuario(t *testing.T) {
	store := &fakeCreateStore{}
	rr := doLobby(t, store, true, map[string]any{"category": "outro", "description": "descricao suficiente"})
	if rr.Code != http.StatusCreated {
		t.Fatalf("esperado 201, veio %d", rr.Code)
	}
	if store.last.UserID == nil || *store.last.UserID != "u1" {
		t.Fatalf("user_id inesperado: %v", store.last.UserID)
	}
	if store.last.Username == nil || *store.last.Username != "ana" {
		t.Fatalf("username inesperado: %v", store.last.Username)
	}
}

func TestFlagHandlerSemUsuario401(t *testing.T) {
	h := &Handler{BugReportStore: &fakeAdminStore{}}
	req := bugJSONRequest(http.MethodPost, "/bugReports/flag", map[string]any{"matchId": "x", "turn": 1})
	rr := httptest.NewRecorder()
	h.BugReportFlagHandler(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401, veio %d", rr.Code)
	}
}

// --- patch ---

type fakePatchStore struct {
	findFn   func(ctx context.Context, id int64) (*models.BugReport, error)
	updateFn func(ctx context.Context, id int64, userID string, cat, desc *string, context json.RawMessage) (*models.BugReport, error)
}

func (f *fakePatchStore) Find(ctx context.Context, id int64) (*models.BugReport, error) {
	return f.findFn(ctx, id)
}

func (f *fakePatchStore) Update(ctx context.Context, id int64, userID string, cat, desc *string, context json.RawMessage) (*models.BugReport, error) {
	return f.updateFn(ctx, id, userID, cat, desc, context)
}
func doPatch(t *testing.T, store bugPatchStore, id int64, userID string, body map[string]any) *httptest.ResponseRecorder {
	t.Helper()
	req := bugJSONRequest(http.MethodPatch, "/bugReports/1", body)
	rr := httptest.NewRecorder()
	patchBugReport(rr, req, id, userID, store)
	return rr
}

func strPtrRemoved() {}

func TestPatchRelatoNaoEncontrado404(t *testing.T) {
	store := &fakePatchStore{findFn: func(ctx context.Context, id int64) (*models.BugReport, error) {
		return nil, gorm.ErrRecordNotFound
	}}
	rr := doPatch(t, store, 1, "u1", map[string]any{"category": "tela"})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, veio %d", rr.Code)
	}
}

func TestPatchRelatoLobby403(t *testing.T) {
	uid := "u1"
	store := &fakePatchStore{findFn: func(ctx context.Context, id int64) (*models.BugReport, error) {
		return &models.BugReport{ID: 1, Source: "lobby", UserID: &uid}, nil
	}}
	rr := doPatch(t, store, 1, "u1", map[string]any{"category": "tela"})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, veio %d", rr.Code)
	}
}

func TestPatchDeOutroUsuario403(t *testing.T) {
	other := "u2"
	store := &fakePatchStore{findFn: func(ctx context.Context, id int64) (*models.BugReport, error) {
		return &models.BugReport{ID: 1, Source: "partida", UserID: &other}, nil
	}}
	rr := doPatch(t, store, 1, "u1", map[string]any{"category": "tela"})
	if rr.Code != http.StatusForbidden {
		t.Fatalf("esperado 403, veio %d", rr.Code)
	}
}

func TestPatchDono200(t *testing.T) {
	uid := "u1"
	cat := "tela"
	report := &models.BugReport{ID: 1, Source: "partida", UserID: &uid, Category: &cat, Turns: []int32{3}}
	store := &fakePatchStore{
		findFn: func(ctx context.Context, id int64) (*models.BugReport, error) { return report, nil },
		updateFn: func(ctx context.Context, id int64, userID string, c, d *string, context json.RawMessage) (*models.BugReport, error) {
			return report, nil
		},
	}
	rr := doPatch(t, store, 1, "u1", map[string]any{"category": "tela", "description": "descricao do bug"})
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}
}

// --- auth opcional ---

func TestOptionalUser(t *testing.T) {
	t.Setenv("JWT_SECRET", "segredo")
	t.Setenv("JWT_EXPIRATION", "1")
	jwtSvc := services.NewJWTService()
	h := &Handler{JWTService: jwtSvc}

	valid, err := jwtSvc.GenerateJWT(structures.UserClaims{Id: "u1", Username: "ana"})
	if err != nil {
		t.Fatalf("generate jwt: %v", err)
	}

	expired := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id": "u1", "username": "ana", "aud": services.LoginAudience,
		"exp": time.Now().Add(-time.Hour).Unix(),
	})
	expiredStr, _ := expired.SignedString([]byte("segredo"))

	cases := []struct {
		name   string
		header string
		want   bool
	}{
		{"ausente", "", false},
		{"invalido", "Bearer lixo", false},
		{"vencido", "Bearer " + expiredStr, false},
		{"valido", "Bearer " + valid, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/bugReports", nil)
			if c.header != "" {
				req.Header.Set("Authorization", c.header)
			}
			user, logged := h.optionalUser(req)
			if logged != c.want {
				t.Fatalf("logged = %v, esperado %v", logged, c.want)
			}
			if c.want && user.Id != "u1" {
				t.Fatalf("id = %q, esperado u1", user.Id)
			}
		})
	}
}
