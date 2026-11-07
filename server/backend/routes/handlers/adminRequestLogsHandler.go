package handlers

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/MatheusGoncalves540/Hoodwink/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	defaultAdminLimit     = 50
	maxAdminLimit         = 200
	maxAdminPage          = 1000
	maxAdminTextLen       = 64
	defaultAdminRange     = 24 * time.Hour
	maxAdminRange         = 31 * 24 * time.Hour
	maxAdminRangeArchived = 183 * 24 * time.Hour
	adminStatementTimeout = "5s"
)

var statusPattern = regexp.MustCompile(`^[1-5]xx$`)

type adminRequestLogFilters struct {
	from      time.Time
	to        time.Time
	route     string
	statusMin *int
	statusMax *int
	userID    string
	username  string
	service   string
	ip        string
	archived  bool
	page      int
	limit     int
}

type adminRequestLogItem struct {
	ID         int64     `json:"id"`
	At         time.Time `json:"at"`
	Service    string    `json:"service"`
	Method     *string   `json:"method"`
	Path       *string   `json:"path"`
	Route      *string   `json:"route"`
	Status     *int      `json:"status"`
	DurationMs *int      `json:"duration_ms"`
	IP         *string   `json:"ip"`
	UserAgent  *string   `json:"user_agent"`
	UserID     *string   `json:"user_id"`
	Username   *string   `json:"username"`
	RequestID  *string   `json:"request_id"`
	Body       *string   `json:"body"`
	Error      *string   `json:"error"`
}

type adminRequestLogsResponse struct {
	Items []adminRequestLogItem `json:"items"`
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// AdminRequestLogsHandler lista as requests para o super-admin.
func (h *Handler) AdminRequestLogsHandler(w http.ResponseWriter, r *http.Request) {
	f, err := parseAdminRequestLogFilters(r.URL.Query())
	if err != nil {
		utils.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	adminRequestLogs(w, r, h.DB, f)
}

func adminRequestLogs(w http.ResponseWriter, r *http.Request, db *gorm.DB, f adminRequestLogFilters) {
	where := buildAdminWhere(f)
	table := "request_logs"
	if f.archived {
		table = "request_logs_archive"
	}

	var total int64
	items := make([]adminRequestLogItem, 0)

	err := db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SET LOCAL statement_timeout = '" + adminStatementTimeout + "'").Error; err != nil {
			return err
		}
		if err := tx.Table(table).Where(where.cond, where.args...).Count(&total).Error; err != nil {
			return err
		}
		return tx.Table(table).
			Where(where.cond, where.args...).
			Order("at DESC").
			Limit(f.limit).
			Offset((f.page - 1) * f.limit).
			Scan(&items).Error
	})
	if err != nil {
		utils.SendError(w, "erro ao consultar logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(adminRequestLogsResponse{
		Items: items,
		Total: total,
		Page:  f.page,
		Limit: f.limit,
	})
}

type adminWhere struct {
	cond string
	args []any
}

// buildAdminWhere monta a cláusula WHERE com placeholders e LIKE escapado.
// Nunca interpola valores de filtro direto no SQL.
func buildAdminWhere(f adminRequestLogFilters) adminWhere {
	conds := make([]string, 0, 8)
	args := make([]any, 0, 8)

	conds = append(conds, "at >= ?", "at <= ?")
	args = append(args, f.from, f.to)

	if f.route != "" {
		conds = append(conds, `route LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(f.route)+"%")
	}
	if f.statusMin != nil && f.statusMax != nil {
		conds = append(conds, "status >= ?", "status <= ?")
		args = append(args, *f.statusMin, *f.statusMax)
	}
	if f.userID != "" {
		conds = append(conds, "user_id = ?")
		args = append(args, f.userID)
	}
	if f.username != "" {
		conds = append(conds, `username LIKE ? ESCAPE '\'`)
		args = append(args, "%"+escapeLike(f.username)+"%")
	}
	if f.service != "" {
		conds = append(conds, "service = ?")
		args = append(args, f.service)
	}
	if f.ip != "" {
		conds = append(conds, "ip = ?")
		args = append(args, f.ip)
	}

	return adminWhere{cond: strings.Join(conds, " AND "), args: args}
}

// escapeLike escapa %, _ e \ para o operador LIKE com ESCAPE '\'.
func escapeLike(s string) string {
	r := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return r.Replace(s)
}

func parseAdminRequestLogFilters(q url.Values) (adminRequestLogFilters, error) {
	f := adminRequestLogFilters{page: 1, limit: defaultAdminLimit}

	f.archived = q.Get("archived") == "true" || q.Get("archived") == "1"

	if err := parseRange(q, &f); err != nil {
		return f, err
	}

	if v := q.Get("route"); v != "" {
		f.route = truncateUTF8Safe(v, maxAdminTextLen)
	}

	if v := q.Get("status"); v != "" {
		min, max, err := parseAdminStatus(v)
		if err != nil {
			return f, err
		}
		f.statusMin = &min
		f.statusMax = &max
	}

	if v := q.Get("userId"); v != "" {
		if _, err := uuid.Parse(v); err != nil {
			return f, fmt.Errorf("userId inválido")
		}
		f.userID = v
	}

	if v := q.Get("username"); v != "" {
		f.username = truncateUTF8Safe(v, maxAdminTextLen)
	}

	if v := q.Get("service"); v != "" {
		if v != "backend" && v != "gameServer" {
			return f, fmt.Errorf("service inválido")
		}
		f.service = v
	}

	if v := q.Get("ip"); v != "" {
		if net.ParseIP(v) == nil {
			return f, fmt.Errorf("ip inválido")
		}
		f.ip = v
	}

	if v := q.Get("page"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 || n > maxAdminPage {
			return f, fmt.Errorf("page inválido")
		}
		f.page = n
	}

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n < 1 {
			n = defaultAdminLimit
		}
		if n > maxAdminLimit {
			n = maxAdminLimit
		}
		f.limit = n
	}

	return f, nil
}

func parseRange(q url.Values, f *adminRequestLogFilters) error {
	hasFrom := q.Has("from")
	hasTo := q.Has("to")
	if hasFrom != hasTo {
		return fmt.Errorf("from e to devem vir juntos")
	}

	if !hasFrom {
		now := time.Now().UTC()
		f.to = now
		f.from = now.Add(-defaultAdminRange)
		return nil
	}

	from, err := time.Parse(time.RFC3339, q.Get("from"))
	if err != nil {
		return fmt.Errorf("from inválido")
	}
	to, err := time.Parse(time.RFC3339, q.Get("to"))
	if err != nil {
		return fmt.Errorf("to inválido")
	}
	if from.After(to) {
		return fmt.Errorf("from não pode ser maior que to")
	}

	max := maxAdminRange
	if f.archived {
		max = maxAdminRangeArchived
	}
	if to.Sub(from) > max {
		return fmt.Errorf("intervalo máximo de %d dias", int(max.Hours()/24))
	}

	f.from = from
	f.to = to
	return nil
}

func parseAdminStatus(s string) (int, int, error) {
	if statusPattern.MatchString(s) {
		d := int(s[0] - '0')
		return d * 100, d*100 + 99, nil
	}
	n, err := strconv.Atoi(s)
	if err != nil || n < 100 || n > 599 {
		return 0, 0, fmt.Errorf("status inválido")
	}
	return n, n, nil
}
