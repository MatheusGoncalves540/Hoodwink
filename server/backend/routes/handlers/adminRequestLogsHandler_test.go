package handlers

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestEscapeLike(t *testing.T) {
	cases := map[string]string{
		"abc":   "abc",
		"100%":  `100\%`,
		"a_b":   `a\_b`,
		`a\b`:   `a\\b`,
		"%_\\%": `\%\_\\\%`,
	}
	for in, want := range cases {
		if got := escapeLike(in); got != want {
			t.Fatalf("escapeLike(%q) = %q, quer %q", in, got, want)
		}
	}
}

func TestParseAdminStatus(t *testing.T) {
	cases := []struct {
		in      string
		wantMin int
		wantMax int
		wantErr bool
	}{
		{"4xx", 400, 499, false},
		{"1xx", 100, 199, false},
		{"200", 200, 200, false},
		{"599", 599, 599, false},
		{"abc", 0, 0, true},
		{"99", 0, 0, true},
		{"600", 0, 0, true},
	}
	for _, c := range cases {
		min, max, err := parseAdminStatus(c.in)
		if c.wantErr {
			if err == nil {
				t.Fatalf("parseAdminStatus(%q) deveria errar", c.in)
			}
			continue
		}
		if err != nil || min != c.wantMin || max != c.wantMax {
			t.Fatalf("parseAdminStatus(%q) = %d,%d,%v; quer %d,%d", c.in, min, max, err, c.wantMin, c.wantMax)
		}
	}
}

func q(values ...string) url.Values {
	v := url.Values{}
	for i := 0; i+1 < len(values); i += 2 {
		v.Set(values[i], values[i+1])
	}
	return v
}

func TestParseAdminRequestLogFiltersDefault24h(t *testing.T) {
	f, err := parseAdminRequestLogFilters(q())
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if f.to.Sub(f.from) != defaultAdminRange {
		t.Fatalf("range padrão = %v, esperado %v", f.to.Sub(f.from), defaultAdminRange)
	}
	if f.page != 1 || f.limit != defaultAdminLimit {
		t.Fatalf("page/limit padrão = %d/%d", f.page, f.limit)
	}
}

func TestParseAdminRequestLogFiltersInvalidos(t *testing.T) {
	now := time.Now().UTC().Format(time.RFC3339)
	future := time.Now().UTC().Add(40 * 24 * time.Hour).Format(time.RFC3339)
	old := time.Now().UTC().Add(-40 * 24 * time.Hour).Format(time.RFC3339)

	cases := []struct {
		name string
		v    url.Values
	}{
		{"status=abc", q("status", "abc")},
		{"userId=x", q("userId", "x")},
		{"service=foo", q("service", "foo")},
		{"ip inválido", q("ip", "999.1.1.1")},
		{"intervalo 40 dias", q("from", old, "to", now)},
		{"from sem to", q("from", old)},
		{"page 0", q("page", "0")},
		{"page 1001", q("page", "1001")},
		{"from > to", q("from", future, "to", now)},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := parseAdminRequestLogFilters(c.v); err == nil {
				t.Fatalf("esperado erro para %s", c.name)
			}
		})
	}
}

func TestParseAdminRequestLogFiltersLimitClamp(t *testing.T) {
	f, err := parseAdminRequestLogFilters(q("limit", "500"))
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}
	if f.limit != maxAdminLimit {
		t.Fatalf("limit=500 deveria virar %d, veio %d", maxAdminLimit, f.limit)
	}
}

func TestParseAdminRequestLogFiltersArchived100DiasOK(t *testing.T) {
	old := time.Now().UTC().Add(-100 * 24 * time.Hour).Format(time.RFC3339)
	now := time.Now().UTC().Format(time.RFC3339)
	if _, err := parseAdminRequestLogFilters(q("from", old, "to", now, "archived", "true")); err != nil {
		t.Fatalf("archived deveria aceitar 100 dias, erro: %v", err)
	}
}

func TestBuildAdminWhereEscapaLike(t *testing.T) {
	now := time.Now().UTC()
	f := adminRequestLogFilters{from: now.Add(-24 * time.Hour), to: now, route: "100%", page: 1, limit: 50}

	w := buildAdminWhere(f)

	if !strings.Contains(w.cond, `ESCAPE '\'`) {
		t.Fatalf("cond deveria usar ESCAPE '\\': %q", w.cond)
	}
	found := false
	for _, a := range w.args {
		if s, ok := a.(string); ok && s == `%100\%%` {
			found = true
		}
	}
	if !found {
		t.Fatalf("route com %% deveria virar %%100\\%%%% nos args: %v", w.args)
	}
}

func TestBuildAdminWhereInjecaoNaoViraWildcard(t *testing.T) {
	now := time.Now().UTC()
	f := adminRequestLogFilters{from: now.Add(-24 * time.Hour), to: now, route: "';--"}

	w := buildAdminWhere(f)

	for _, a := range w.args {
		if s, ok := a.(string); ok && s == `%';--%` {
			// O ponto-e-vírgula e comentário ficam como texto literal, sem wildcard.
			if !strings.Contains(w.cond, `ESCAPE '\'`) {
				t.Fatalf("falta ESCAPE para o LIKE literal: %q", w.cond)
			}
			return
		}
	}
	t.Fatalf("route ';-- não apareceu como literal nos args: %v", w.args)
}
