package magefiles

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/internal/preservation"
)

func pinnedCorpusFixture(t *testing.T) string {
	t.Helper()
	t.Chdir(t.TempDir())
	const input = `[{"name":"Bear"}]`
	digest, err := preservation.Digest(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	pin := preservation.Pin{
		DownloadURI:    "https://example.invalid/oracle.jsonl.gz",
		DownloadSHA256: strings.Repeat("a", 64), JSONSHA256: digest,
		UpdatedAt: "2026-10-13", RulesEdition: "2026-09-25", BaselineRevision: "baseline",
	}
	if err := os.MkdirAll(filepath.Dir(corpusPinPath), 0o750); err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(pin)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(corpusPinPath, data, 0o600); err != nil {
		t.Fatal(err)
	}
	path := "oracle.json"
	if err := os.WriteFile(path, []byte(input), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("COUNCIL4_ORACLE_CARDS", path)
	return path
}

func TestPinnedCorpusUsesVerifiedLocalInput(t *testing.T) {
	path := pinnedCorpusFixture(t)
	got, err := pinnedOracleCards()
	if err != nil {
		t.Fatal(err)
	}
	if got != path {
		t.Fatalf("corpus = %q, want %q", got, path)
	}
	if err := os.WriteFile(path, []byte("different bytes"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := pinnedOracleCards(); err == nil {
		t.Fatal("accepted modified local input")
	}
}

func TestPinnedCorpusDoesNotDownloadMissingInput(t *testing.T) {
	path := pinnedCorpusFixture(t)
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if _, err := pinnedOracleCards(); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing corpus error = %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing local input was created: %v", err)
	}
}
