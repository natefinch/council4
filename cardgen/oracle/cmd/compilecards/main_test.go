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
		{"id":"v","oracle_id":"ov","name":"Vanilla Bear","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature — Bear","power":"2","toughness":"2"},
		{"id":"k","oracle_id":"ok","name":"Flying Bear","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature — Bear","oracle_text":"Flying","power":"2","toughness":"2"},
		{"id":"u","oracle_id":"ou","name":"Drawing Bear","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature — Bear","oracle_text":"When this creature enters, draw a card, then discard a card.","power":"2","toughness":"2"}
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

func TestRunExcludesCardsOutsideCorpusPolicy(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "oracle.json")
	reportPath := filepath.Join(directory, "report.json")
	corpus := `[
		{"id":"paper","name":"Paper Card","layout":"normal","games":["paper"],"legalities":{"legacy":"banned"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"funny","name":"Legal Funny Card","layout":"normal","set_type":"funny","games":["paper"],"legalities":{"legacy":"legal"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"token","name":"Bear","layout":"token","set_type":"token","games":["paper"],"type_line":"Creature — Bear","power":"2","toughness":"2"},
		{"id":"digital-print","name":"Paper Identity Digital Printing","layout":"normal","set_type":"masters","games":["mtgo"],"digital":true,"legalities":{"legacy":"legal"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"alchemy","name":"Alchemy Card","layout":"normal","set_type":"alchemy","games":["arena"],"legalities":{"legacy":"legal"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"digital","name":"Digital Card","layout":"normal","set_type":"expansion","games":["arena"],"legalities":{"historic":"legal"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"memorabilia","name":"Challenge Card","layout":"normal","set_type":"memorabilia","games":["paper"],"type_line":"Sorcery"},
		{"id":"unset","name":"Illegal Funny Card","layout":"normal","set_type":"funny","games":["paper"],"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"scheme","name":"Scheme Card","layout":"scheme","set_type":"archenemy","games":["paper"],"legalities":{"legacy":"legal"},"type_line":"Scheme"},
		{"id":"minigame","name":"Minigame Card","layout":"token","set_type":"minigame","games":["paper"],"type_line":"Card"}
	]`
	if err := os.WriteFile(input, []byte(corpus), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(config{
		inputPath:  input,
		outputRoot: filepath.Join(directory, "cards"),
		reportPath: reportPath,
		format:     "json",
		workers:    2,
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		`"card_count": 10`,
		`"eligible_count": 4`,
		`"generated_count": 4`,
		`"excluded_count": 6`,
		`"reason": "alchemy"`,
		`"reason": "digital-only"`,
		`"reason": "memorabilia"`,
		`"reason": "no-sanctioned-paper-legality"`,
		`"reason": "special-format"`,
	} {
		if !strings.Contains(string(data), wanted) {
			t.Errorf("report missing %q:\n%s", wanted, data)
		}
	}
}

func TestWriteTextReportListsEachExclusionOnce(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "report.txt")
	output := report{
		CardCount:     2,
		EligibleCount: 1,
		ExcludedCount: 1,
		Excluded: []excluded{{
			Name:   "Excluded Card",
			Reason: cardgen.ExcludeMemorabilia,
		}},
	}
	if err := writeReport(path, "text", output); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Count(string(data), "Excluded Card\texcluded\tmemorabilia"); got != 1 {
		t.Fatalf("exclusion count = %d, report:\n%s", got, data)
	}
}

func TestCompileCorpusRejectsPathCollisions(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"one","name":"Same Name","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature"},
		{"id":"two","name":"Same Name","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature"}
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
