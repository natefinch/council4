package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
		{"id":"u","oracle_id":"ou","name":"Drawing Bear","layout":"normal","type_line":"Creature — Bear","oracle_text":"When this creature enters, draw a card.","power":"2","toughness":"2"}
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

func TestCompileCorpusRejectsPathCollisions(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"one","name":"Same Name","layout":"normal","type_line":"Creature"},
		{"id":"two","name":"Same Name","layout":"normal","type_line":"Creature"}
	]`
	results, err := compileCorpus(strings.NewReader(input), 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, result := range results {
		if len(result.diagnostics) == 0 || result.diagnostics[0].Summary != "generated path collision" {
			t.Fatalf("result = %#v", result)
		}
	}
}

func TestSafeFileNameAvoidsGoBuildSuffix(t *testing.T) {
	t.Parallel()
	if got := safeFileName("Bayou Dragonfly"); got != "bayou_dragonfly_card" {
		t.Fatalf("safeFileName = %q", got)
	}
	if got := safeFileName("Dragonfly"); got != "dragonfly_card" {
		t.Fatalf("safeFileName = %q", got)
	}
	if got := safeFileName("Dragonfly Hatchling"); got != "dragonfly_hatchling" {
		t.Fatalf("safeFileName = %q", got)
	}
	if got := safeFileName("Memory Test"); got != "memory_test_card" {
		t.Fatalf("safeFileName = %q", got)
	}
	if got := safeFileName("Cards"); got != "cards_card" {
		t.Fatalf("safeFileName = %q", got)
	}
}
