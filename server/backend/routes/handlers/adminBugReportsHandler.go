package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultBugLimit = 50
	maxBugLimit     = 200
)

var bugStatuses = []string{"novo", "visto", "resolvido"}

func validBugStatus(s string) bool {
	for _, v := range bugStatuses {
		if s == v {
			return true
		}
	}
	return false
}

type bugListFilter struct {
	status  string
	source  string
	matchID string
	page    int
	limit   int
}

type adminBugReportsResponse struct {
	Items []bugReportDTO `json:"items"`
	Total int64          `json:"total"`
	Novos int64          `json:"novos"`
	Page  int            `json:"page"`
	Limit int            `json:"limit"`
}

// AdminBugReportsListHandler lista relatos para o super-admin.
func (h *Handler) AdminBugReportsListHandler(w http.ResponseWriter, r *http.Request) {
	f, err := parseBugListFilters(r.URL.Query())
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	listAdminBugReports(w, r, h.BugReportStore, f)
}

func listAdminBugReports(w http.ResponseWriter, r *http.Request, store bugAdminStore, f bugListFilter) {
	items, total, novos, err := store.List(r.Context(), f)
	if err != nil {
		utils.SendError(w, "erro ao consultar relatos", http.StatusInternalServerError)
		return
	}
	dtos := make([]bugReportDTO, 0, len(items))
	for _, it := range items {
		dtos = append(dtos, toBugReportDTO(it))
	}
	writeBugJSON(w, http.StatusOK, adminBugReportsResponse{
		Items: dtos,
		Total: total,
		Novos: novos,
		Page:  f.page,
		Limit: f.limit,
	})
}

// AdminBugReportsPatchHandler troca o status de um relato (super-admin).
func (h *Handler) AdminBugReportsPatchHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil || id <= 0 {
		utils.SendError(w, "id inválido", http.StatusBadRequest)
		return
	}
	var req struct {
		Status string `json:"status"`
	}
	if !decodeBugBody(w, r, &req) {
		return
	}
	if !validBugStatus(req.Status) {
		utils.SendError(w, "status inválido", http.StatusBadRequest)
		return
	}
	patchAdminBugReport(w, r, id, req.Status, h.BugReportStore)
}

func patchAdminBugReport(w http.ResponseWriter, r *http.Request, id int64, status string, store bugAdminStore) {
	report, err := store.UpdateStatus(r.Context(), id, status)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			utils.SendError(w, "relato não encontrado", http.StatusNotFound)
			return
		}
		utils.SendError(w, "erro ao atualizar relato", http.StatusInternalServerError)
		return
	}
	writeBugJSON(w, http.StatusOK, toBugReportDTO(*report))
}

func parseBugListFilters(q url.Values) (bugListFilter, error) {
	f := bugListFilter{page: 1, limit: defaultBugLimit}

	if v := q.Get("status"); v != "" {
		if !validBugStatus(v) {
			return f, errors.New("status inválido")
		}
		f.status = v
	}
	if v := q.Get("source"); v != "" {
		if v != "partida" && v != "lobby" {
			return f, errors.New("source inválido")
		}
		f.source = v
	}
	if v := q.Get("matchId"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return f, errors.New("matchId inválido")
		}
		f.matchID = v
	}
	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			return f, errors.New("page inválido")
		}
		f.page = n
	}
	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			n = defaultBugLimit
		}
		if n > maxBugLimit {
			n = maxBugLimit
		}
		f.limit = n
	}
	return f, nil
}
