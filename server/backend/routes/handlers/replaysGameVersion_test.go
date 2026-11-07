package handlers

import (
	"net/http"
	"strings"
	"testing"
)

func TestSanitizeGameVersion(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"0.1.0", "0.1.0"},
		{" 1.2.3-rc.1 ", "1.2.3-rc.1"},
		{"dev", "dev"},
		{"", "unknown"},
		{"com espaço", "unknown"},
		{"<script>", "unknown"},
		{strings.Repeat("a", 33), "unknown"},
	}
	for _, c := range cases {
		if got := sanitizeGameVersion(c.in); got != c.want {
			t.Errorf("sanitizeGameVersion(%q) = %q, esperado %q", c.in, got, c.want)
		}
	}
}

func TestReplayUploadGravaGameVersion(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	body := uploadBody(buildReplay(t, 1, true), []string{testPlayerA, testPlayerB}, "r1")
	body["game_version"] = " 0.1.0 "

	rr := postUpload(store, signer, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}
	got := store.inserted[testMatchUUID.String()]
	if got == nil {
		t.Fatal("replay não gravado")
	}
	if got.GameVersion != "0.1.0" {
		t.Fatalf("game_version = %q, esperado %q", got.GameVersion, "0.1.0")
	}
}

func TestReplayUploadGameVersionInvalidaViraUnknown(t *testing.T) {
	store := newFakeReplayStore()
	signer, _ := testSigner(t)

	body := uploadBody(buildReplay(t, 1, true), []string{testPlayerA, testPlayerB}, "r1")
	body["game_version"] = "<script>alert(1)</script>"

	rr := postUpload(store, signer, body)
	if rr.Code != http.StatusOK {
		t.Fatalf("esperado 200, veio %d: %s", rr.Code, rr.Body.String())
	}
	got := store.inserted[testMatchUUID.String()]
	if got == nil {
		t.Fatal("replay não gravado")
	}
	if got.GameVersion != "unknown" {
		t.Fatalf("game_version = %q, esperado %q", got.GameVersion, "unknown")
	}
}
