package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEngineWritesSupportedListAndInspectionManifest(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	corpusPath := writeFixture(t, dir, "corpus.json", `[
		{"id":"a","name":"Alpha","layout":"normal","oracle_text":"Flying"},
		{"id":"b","name":"Beta","layout":"normal","oracle_text":"Vigilance"},
		{"id":"g","name":"Gamma","layout":"normal","oracle_text":"Unsupported"},
		{"id":"art","name":"Alpha Art Card","layout":"art_series"}
	]`)
	baselinePath := writeFixture(t, dir, "baseline.json", `{
		"card_count":4,"generated_count":1,"unsupported_count":3,
		"unsupported":[
			{"id":"b","name":"Beta","diagnostics":[{"summary":"unsupported static ability"}]},
			{"id":"g","name":"Gamma","diagnostics":[{"summary":"unsupported Oracle construct"}]},
			{"id":"art","name":"Alpha Art Card","layout":"art_series","diagnostics":[{"summary":"unsupported card layout"}]}
		]
	}`)
	currentPath := writeFixture(t, dir, "current.json", `{
		"card_count":3,"generated_count":2,"unsupported_count":1,
		"unsupported":[
			{"id":"g","name":"Gamma","diagnostics":[{"summary":"unsupported static ability"}]}
		]
	}`)
	generatedRoot := filepath.Join(dir, "generated")
	if err := os.MkdirAll(filepath.Join(generatedRoot, "b"), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(generatedRoot, "b", "beta.go"), []byte("package b"), 0o600); err != nil {
		t.Fatal(err)
	}
	supportedPath := filepath.Join(dir, "supported.md")
	manifestPath := filepath.Join(dir, "manifest.json")

	engine := Engine{Config: Config{
		CorpusPath:     corpusPath,
		BaselineReport: baselinePath,
		CurrentReport:  currentPath,
		GeneratedRoot:  generatedRoot,
		SupportedPath:  supportedPath,
		ManifestPath:   manifestPath,
	}}
	if err := engine.Run(); err != nil {
		t.Fatal(err)
	}

	supported, err := os.ReadFile(supportedPath)
	if err != nil {
		t.Fatal(err)
	}
	if got, want := string(supported), "# Supported Cards\n\nCards supported: 2\n\n- Alpha\n- Beta\n"; got != want {
		t.Fatalf("supported.md:\n%s\nwant:\n%s", got, want)
	}

	var manifest Manifest
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.CardCount != 3 || manifest.GeneratedDelta != 1 || len(manifest.NewlySupported) != 1 {
		t.Fatalf("manifest = %#v", manifest)
	}
	card := manifest.NewlySupported[0]
	if card.ID != "b" || card.Name != "Beta" || card.OracleText != "Vigilance" ||
		card.GeneratedPath != filepath.Join(generatedRoot, "b", "beta.go") {
		t.Fatalf("new card = %#v", card)
	}
	if len(manifest.DiagnosticChanges) != 1 ||
		manifest.DiagnosticChanges[0].Summary != "unsupported Oracle construct" ||
		manifest.DiagnosticChanges[0].Delta != -1 {
		t.Fatalf("diagnostic changes = %#v", manifest.DiagnosticChanges)
	}
	if len(manifest.ChangedDiagnostics) != 1 ||
		manifest.ChangedDiagnostics[0].Name != "Gamma" {
		t.Fatalf("changed diagnostics = %#v", manifest.ChangedDiagnostics)
	}
}

func TestEngineRejectsMissingGeneratedSource(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	corpusPath := writeFixture(t, dir, "corpus.json", `[{"id":"a","name":"Alpha","layout":"normal"}]`)
	baselinePath := writeFixture(t, dir, "baseline.json", `{
		"card_count":1,"generated_count":0,"unsupported_count":1,
		"unsupported":[{"id":"a","name":"Alpha","diagnostics":[{"summary":"unsupported"}]}]
	}`)
	currentPath := writeFixture(t, dir, "current.json", `{
		"card_count":1,"generated_count":1,"unsupported_count":0,"unsupported":[]
	}`)
	err := (&Engine{Config: Config{
		CorpusPath:     corpusPath,
		BaselineReport: baselinePath,
		CurrentReport:  currentPath,
		GeneratedRoot:  filepath.Join(dir, "generated"),
		SupportedPath:  filepath.Join(dir, "supported.md"),
		ManifestPath:   filepath.Join(dir, "manifest.json"),
	}}).Run()
	if err == nil || !strings.Contains(err.Error(), "generated source for Alpha") {
		t.Fatalf("error = %v", err)
	}
}

func writeFixture(t *testing.T, dir, name, contents string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}
