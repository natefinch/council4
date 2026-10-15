package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/internal/preservation"
)

func TestInstallPreservesImplementationsAndFilenameAliases(t *testing.T) {
	root := t.TempDir()
	t.Chdir(root)
	pin := preservation.Pin{
		DownloadURI: "https://example.invalid/pinned.gz", DownloadSHA256: strings.Repeat("1", 64),
		JSONSHA256: strings.Repeat("2", 64), UpdatedAt: "2026-10-13",
		RulesEdition: "2026-09-25", BaselineRevision: "baseline",
	}
	writeTestFile(t, "testdata/card-preservation/corpus.json", mustJSON(t, pin))
	source := "package a\nimport \"github.com/natefinch/council4/mtg/game\"\nvar Accent = newAccent\nfunc newAccent() *game.CardDef { return &game.CardDef{CardFace: game.CardFace{Name:\"Accent\"}} }\n"
	added := strings.ReplaceAll(source, "Accent", "Added")
	original := strings.ReplaceAll(source, `"Accent"`, `"Preserved Accent"`)
	writeTestFile(t, "generated/a/áccent.go", []byte(source))
	writeTestFile(t, "generated/a/added.go", []byte(added))
	writeTestFile(t, "cards/a/accent.go", []byte(original))
	writeTestFile(t, "cards/authored/protected.go", []byte("package authored\n"))
	manifest := preservation.Manifest{
		SchemaVersion: 1, Input: pin,
		Compiler: []preservation.Record{
			{Key: "accent", Name: "Accent", Layout: "normal", DefinitionSHA256: hash([]byte(source))},
			{Key: "added", Name: "Added", Layout: "normal", DefinitionSHA256: hash([]byte(added))},
		},
	}
	writeTestFile(t, "inventory.json", mustJSON(t, manifest))
	if err := run("generated", "inventory.json", "cards", "installed.json"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile("cards/a/accent.go")
	if err != nil || string(got) != original {
		t.Fatal("existing behavior was overwritten", err)
	}
	if _, err := os.Stat("cards/a/áccent.go"); !os.IsNotExist(err) {
		t.Fatal("filename alias redeclared the existing constructor")
	}
	if got, err := os.ReadFile("cards/a/added.go"); err != nil || string(got) != added {
		t.Fatal("missing generated definition", err)
	}
	if got, err := os.ReadFile("cards/authored/protected.go"); err != nil || string(got) != "package authored\n" {
		t.Fatal("authored directory was modified", err)
	}
	data, err := os.ReadFile("installed.json")
	if err != nil {
		t.Fatal(err)
	}
	var coverage installation
	if err := json.Unmarshal(data, &coverage); err != nil {
		t.Fatal(err)
	}
	if len(coverage.Cards) != 2 || coverage.Cards[0].Path != "a/accent.go" ||
		coverage.Cards[0].SHA256 != hash([]byte(original)) {
		t.Fatalf("missing exact coverage of retained definition: %+v", coverage)
	}
	if err := run("generated", "inventory.json", "cards", "installed.json"); err != nil {
		t.Fatal("repeat import failed", err)
	}
	repeated, err := os.ReadFile("installed.json")
	if err != nil || string(repeated) != string(data) {
		t.Fatal("repeat import changed installed coverage", err)
	}
}

func TestCardDirectoryRejectsAuthoredAndTraversal(t *testing.T) {
	for _, directory := range []string{"authored", "../a", "tokens", "tokens/../a", "a/extra", "ab"} {
		if cardDirectory(directory) {
			t.Errorf("unexpected generated directory accepted: %s", directory)
		}
	}
	for _, directory := range []string{"a", "z", "tokens/b"} {
		if !cardDirectory(directory) {
			t.Errorf("valid card directory rejected: %s", directory)
		}
	}
}

func writeTestFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func hash(data []byte) string { return fmt.Sprintf("%x", sha256.Sum256(data)) }
