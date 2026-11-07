package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/db/models"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func TestValidBugStatus(t *testing.T) {
	for _, s := range []string{"novo", "visto", "resolvido"} {
		if !validBugStatus(s) {
			t.Fatalf("status %q deveria ser válido", s)
		}
	}
	if validBugStatus("") || validBugStatus("fechado") {
		t.Fatal("status inválido deveria falhar")
	}
}

func bugQ(values ...string) url.Values {
	v := url.Values{}
	for i := 0; i+1 < len(values); i += 2 {
		v.Set(values[i], values[i+1])
	}
	return v
}

func TestParseBugListFiltersPadrao(t *testing.T) {
	f, err := parseBugListFilters(bugQ())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if f.page != 1 || f.limit != defaultBugLimit {
		t.Fatalf("page/limit padrão = %d/%d", f.page, f.limit)
	}
}

func TestParseBugListFiltersInvalidos(t *testing.T) {
	cases := []url.Values{
		bugQ("status", "fechado"),
		bugQ("source", "x"),
		bugQ("matchId", "nao-uuid"),
		bugQ("page", "0"),
	}
	for _, v := range cases {
		if _, err := parseBugListFilters(v); err == nil {
			t.Fatalf("esperado erro para %v", v)
		}
	}
}

func TestParseBugListFiltersLimitClamp(t *testing.T) {
	f, err := parseBugListFilters(bugQ("limit", "500"))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if f.limit != maxBugLimit {
		t.Fatalf("limit=500 deveria virar %d, veio %d", maxBugLimit, f.limit)
	}
}

type fakeAdminStore struct {
	listFn         func(ctx context.Context, f bugListFilter) ([]models.BugReport, int64, int64, error)
	updateStatusFn func(ctx context.Context, id int64, status string) (*models.BugReport, error)
}

func (f *fakeAdminStore) List(ctx context.Context, fl bugListFilter) ([]models.BugReport, int64, int64, error) {
	return f.listFn(ctx, fl)
}

func (f *fakeAdminStore) Find(ctx context.Context, id int64) (*models.BugReport, error) {
	return nil, gorm.ErrRecordNotFound
}

func (f *fakeAdminStore) UpdateStatus(ctx context.Context, id int64, status string) (*models.BugReport, error) {
	return f.updateStatusFn(ctx, id, status)
}

// Métodos abaixo só existem para satisfazer a interface bugReportStore completa
// (o campo Handler.BugReportStore); não são usados nos testes de admin.
func (f *fakeAdminStore) CountPartida24h(ctx context.Context, userID string) (int64, error) {
	return 0, nil
}

func (f *fakeAdminStore) UpsertTurn(ctx context.Context, userID, username, matchID, ip, ua string, turn int32) (int64, []int32, bool, error) {
	return 0, nil, false, nil
}

func (f *fakeAdminStore) Create(ctx context.Context, report *models.BugReport) error {
	return nil
}

func (f *fakeAdminStore) Update(ctx context.Context, id int64, userID string, cat, desc *string, context json.RawMessage) (*models.BugReport, error) {
	return nil, gorm.ErrRecordNotFound
}

func TestListAdminBugReportsResposta(t *testing.T) {
	uid := "u1"
	store := &fakeAdminStore{listFn: func(ctx context.Context, f bugListFilter) ([]models.BugReport, int64, int64, error) {
		return []models.BugReport{{ID: 1, Source: "partida", UserID: &uid, Turns: []int32{2}}}, 1, 1, nil
	}}
	req := httptest.NewRequest(http.MethodGet, "/admin/bugReports", nil)
	rr := httptest.NewRecorder()
	listAdminBugReports(rr, req, store, bugListFilter{page: 1, limit: 50})

	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), `"total":1`) || !strings.Contains(rr.Body.String(), `"novos":1`) {
		t.Fatalf("resposta inesperada: %s", rr.Body.String())
	}
}

func adminPatchReq(id, body string) *http.Request {
	req := httptest.NewRequest(http.MethodPatch, "/admin/bugReports/"+id, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", id)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestAdminBugReportsPatchStatusInvalido400(t *testing.T) {
	h := &Handler{BugReportStore: &fakeAdminStore{}}
	rr := httptest.NewRecorder()
	h.AdminBugReportsPatchHandler(rr, adminPatchReq("1", `{"status":"fechado"}`))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("esperado 400, veio %d", rr.Code)
	}
}

func TestAdminBugReportsPatchIdInexistente404(t *testing.T) {
	store := &fakeAdminStore{updateStatusFn: func(ctx context.Context, id int64, status string) (*models.BugReport, error) {
		return nil, gorm.ErrRecordNotFound
	}}
	h := &Handler{BugReportStore: store}
	rr := httptest.NewRecorder()
	h.AdminBugReportsPatchHandler(rr, adminPatchReq("1", `{"status":"resolvido"}`))
	if rr.Code != http.StatusNotFound {
		t.Fatalf("esperado 404, veio %d", rr.Code)
	}
}

func TestAdminBugReportsPatchOK(t *testing.T) {
	uid := "u1"
	store := &fakeAdminStore{updateStatusFn: func(ctx context.Context, id int64, status string) (*models.BugReport, error) {
		return &models.BugReport{ID: id, Source: "partida", UserID: &uid, Status: status, CreatedAt: time.Now(), UpdatedAt: time.Now(), Turns: []int32{1}}, nil
	}}
	h := &Handler{BugReportStore: store}
	rr := httptest.NewRecorder()
	h.AdminBugReportsPatchHandler(rr, adminPatchReq("1", `{"status":"resolvido"}`))
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}
}
