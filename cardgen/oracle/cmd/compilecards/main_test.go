package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
)

func TestRunGeneratesOnlyFullySupportedCards(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "oracle.json")
	output := filepath.Join(directory, "cards")
	reportPath := filepath.Join(directory, "report.json")
	corpus := `[
		{"id":"v","oracle_id":"ov","name":"Vanilla Bear","layout":"normal","type_line":"Creature — Bear","power":"2","toughness":"2"},
		{"id":"k","oracle_id":"ok","name":"Flying Bear","layout":"normal","type_line":"Creature — Bear","oracle_text":"Flying","power":"2","toughness":"2"},
		{"id":"u","oracle_id":"ou","name":"Drawing Bear","layout":"normal","type_line":"Creature — Bear","oracle_text":"When this creature enters, draw a card, then discard a card.","power":"2","toughness":"2"}
	]`
	if err := os.WriteFile(input, []byte(corpus), 0o600); err != nil {
		t.Fatal(err)
	}
	err := run(config{
		inputPath:  input,
		outputRoot: output,
		reportPath: reportPath,
		format:     "json",
		workers:    3,
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, relative := range []string{
		filepath.Join("v", "vanilla_bear.go"),
		filepath.Join("f", "flying_bear.go"),
		filepath.Join("v", "cards.go"),
		filepath.Join("f", "cards.go"),
	} {
		if _, err := os.Stat(filepath.Join(output, relative)); err != nil {
			t.Errorf("%s was not generated: %v", relative, err)
		}
	}
	if _, err := os.Stat(filepath.Join(output, "d", "drawing_bear.go")); !os.IsNotExist(err) {
		t.Fatalf("unsupported card file exists or stat failed: %v", err)
	}
	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		`"card_count": 3`,
		`"generated_count": 2`,
		`"unsupported_count": 1`,
		`"name": "Drawing Bear"`,
	} {
		if !strings.Contains(string(report), wanted) {
			t.Errorf("report missing %q:\n%s", wanted, report)
		}
	}
}

func TestCompileCorpusDisambiguatesPathAndIdentifierCollisions(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"one","oracle_id":"oracle-one","name":"Same Name","layout":"normal","type_line":"Creature"},
		{"id":"two","oracle_id":"oracle-two","name":"Same Name","layout":"normal","type_line":"Creature"}
	]`
	results, err := compileCorpus(strings.NewReader(input), 2)
	if err != nil {
		t.Fatal(err)
	}
	wantedPaths := map[string]bool{
		filepath.Join("s", "same_name.go"):                   true,
		filepath.Join("s", "same_name_scryfalloracletwo.go"): true,
	}
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) != 0 {
			t.Fatalf("result = %#v", result)
		}
		if !wantedPaths[result.relative] {
			t.Fatalf("relative path = %q", result.relative)
		}
		if result.card.OracleID == "oracle-one" && !strings.Contains(result.source, "var SameName =") {
			t.Fatalf("canonical source lacks stable identifier:\n%s", result.source)
		}
		if result.card.OracleID == "oracle-two" && !strings.Contains(result.source, "var SameNameScryfalloracletwo =") {
			t.Fatalf("colliding source lacks disambiguated identifier:\n%s", result.source)
		}
		if !strings.Contains(result.source, `"Same Name"`) {
			t.Fatalf("source changed printed name:\n%s", result.source)
		}
	}
}

func TestWriteSupportedRemovesSupersededIdentityPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	oldRelative := filepath.Join("s", "same_name.go")
	oldPath := filepath.Join(root, oldRelative)
	if err := os.MkdirAll(filepath.Dir(oldPath), 0o750); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldPath, []byte("package s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	results := []result{{
		card:       cardgen.ScryfallCard{Name: "Same Name"},
		relative:   filepath.Join("s", "same_name_scryfalltwo.go"),
		superseded: oldRelative,
		source:     "package s\n\nimport \"github.com/natefinch/council4/mtg/game\"\n\nvar SameNameScryfalltwo = &game.CardDef{}\n",
	}}

	if err := writeSupported(root, results); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("superseded path still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, results[0].relative)); err != nil {
		t.Fatalf("disambiguated source missing: %v", err)
	}
}

func TestWriteSupportedRemovesObsoleteSuffixedIdentityPath(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	directory := filepath.Join(root, "s")
	if err := os.MkdirAll(directory, 0o750); err != nil {
		t.Fatal(err)
	}
	obsolete := filepath.Join(directory, "same_name_scryfallold.go")
	if err := os.WriteFile(obsolete, []byte("package s\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	results := []result{{
		card:     cardgen.ScryfallCard{Name: "Same Name"},
		relative: filepath.Join("s", "same_name.go"),
		source:   "package s\n\nimport \"github.com/natefinch/council4/mtg/game\"\n\nvar SameName = &game.CardDef{}\n",
	}}

	if err := writeSupported(root, results); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(obsolete); !os.IsNotExist(err) {
		t.Fatalf("obsolete identity path still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, results[0].relative)); err != nil {
		t.Fatalf("canonical source missing: %v", err)
	}
}

func TestSafeFileNameAvoidsGoBuildSuffix(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"Bayou Dragonfly":     "bayou_dragonfly_card",
		"Dragonfly":           "dragonfly_card",
		"Dragonfly Hatchling": "dragonfly_hatchling",
		"Memory Test":         "memory_test_card",
		"Cards":               "cards_card",
	}
	for name, want := range tests {
		if got := cardgen.CardNameToSafeFileName(name); got != want {
			t.Fatalf("CardNameToSafeFileName(%q) = %q, want %q", name, got, want)
		}
	}
}
