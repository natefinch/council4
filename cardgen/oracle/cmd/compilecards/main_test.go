package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/cardgen/oracle/shared"
)

func TestRunGeneratesOnlyFullySupportedCards(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "compiler.json")
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
		filepath.Join("d", "drawing_bear.go"),
		filepath.Join("v", "cards.go"),
		filepath.Join("f", "cards.go"),
		filepath.Join("d", "cards.go"),
	} {
		if _, err := os.Stat(filepath.Join(output, relative)); err != nil {
			t.Errorf("%s was not generated: %v", relative, err)
		}
	}
	report, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		`"card_count": 3`,
		`"generated_count": 3`,
		`"unsupported_count": 0`,
	} {
		if !strings.Contains(string(report), wanted) {
			t.Errorf("report missing %q:\n%s", wanted, report)
		}
	}
}

func TestRunReportsCompilerAndBackendDiagnostics(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "compiler.json")
	reportPath := filepath.Join(directory, "report.json")
	corpus := `[
		{
			"id":"u",
			"oracle_id":"ou",
			"name":"Multiply Unsupported Bear",
			"layout":"normal",
			"games":["paper"],
			"legalities":{"commander":"legal"},
			"type_line":"Creature — Bear",
			"oracle_text":"Dance: Draw a card.\nWhenever this creature attacks the player with the most life, you gain 1 life.",
			"power":"2",
			"toughness":"2"
		}
	]`
	if err := os.WriteFile(input, []byte(corpus), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := run(config{
		inputPath:  input,
		outputRoot: filepath.Join(directory, "cards"),
		reportPath: reportPath,
		format:     "json",
		workers:    1,
	}); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var got report
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.Unsupported) != 1 {
		t.Fatalf("unsupported cards = %d, want 1", len(got.Unsupported))
	}
	summaries := make(map[string]bool)
	for _, diagnostic := range got.Unsupported[0].Diagnostics {
		summaries[diagnostic.Summary] = true
	}
	for _, want := range []string{
		"unsupported cost",
		"unsupported activation cost",
		"unsupported triggered ability",
	} {
		if !summaries[want] {
			t.Errorf("diagnostics missing %q: %#v", want, got.Unsupported[0].Diagnostics)
		}
	}
}

func TestRunExcludesCardsOutsideCorpusPolicy(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	input := filepath.Join(directory, "compiler.json")
	reportPath := filepath.Join(directory, "report.json")
	corpus := `[
		{"id":"paper","name":"Paper Card","layout":"normal","games":["paper"],"legalities":{"legacy":"banned"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"funny","name":"Legal Funny Card","layout":"normal","set_type":"funny","games":["paper"],"legalities":{"legacy":"legal"},"type_line":"Creature","power":"1","toughness":"1"},
		{"id":"token","oracle_id":"11111111-1111-1111-1111-111111111111","name":"Bear","layout":"token","set_type":"token","games":["paper"],"type_line":"Creature — Bear","power":"2","toughness":"2"},
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

func TestWriteSupportDocumentation(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	readmePath := filepath.Join(directory, "README.md")
	if err := os.WriteFile(
		readmePath,
		[]byte("# Test\n\n<!-- card-support:start -->\nold\n<!-- card-support:end -->\n"),
		0o600,
	); err != nil {
		t.Fatal(err)
	}
	output := report{
		CardCount:        4,
		EligibleCount:    3,
		GeneratedCount:   2,
		UnsupportedCount: 1,
		ExcludedCount:    1,
		Unsupported: []unsupported{{
			Name: "Unsupported [Card]",
			Diagnostics: []reportDiagnostic{{
				Summary: "unsupported ability",
				Detail:  "cannot handle\nthis",
			}},
		}},
	}
	results := []result{
		{card: cardgen.ScryfallCard{Name: "Zulu"}},
		{card: cardgen.ScryfallCard{Name: "alpha"}},
		{
			card:        cardgen.ScryfallCard{Name: "Unsupported [Card]"},
			diagnostics: []shared.Diagnostic{{Summary: "unsupported ability"}},
		},
		{
			card:      cardgen.ScryfallCard{Name: "Excluded"},
			exclusion: cardgen.ExcludeDigitalOnly,
		},
	}
	cfg := config{
		supportedPath:          filepath.Join(directory, "supported.md"),
		unsupportedPath:        filepath.Join(directory, "unsupported.md"),
		unsupportedReasonsPath: filepath.Join(directory, "unsupported-reasons.md"),
		readmePath:             readmePath,
	}
	if err := writeSupportDocumentation(cfg, output, results); err != nil {
		t.Fatal(err)
	}
	supported, err := os.ReadFile(cfg.supportedPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		"**2 of 3 cards eligible for paper support (66.7%)**",
		"1 additional digital, special-format",
		"- alpha\n- Zulu\n",
	} {
		if !strings.Contains(string(supported), wanted) {
			t.Errorf("supported.md missing %q:\n%s", wanted, supported)
		}
	}
	unsupportedData, err := os.ReadFile(cfg.unsupportedPath)
	if err != nil {
		t.Fatal(err)
	}
	if wanted := "- **Unsupported \\[Card\\]** — unsupported ability: cannot handle this"; !strings.Contains(string(unsupportedData), wanted) {
		t.Errorf("unsupported.md missing %q:\n%s", wanted, unsupportedData)
	}
	reasons, err := os.ReadFile(cfg.unsupportedReasonsPath)
	if err != nil {
		t.Fatal(err)
	}
	if wanted := "| 1 | unsupported ability | 1 | 1 | 100.0% | - |"; !strings.Contains(string(reasons), wanted) {
		t.Errorf("unsupported-reasons.md missing %q:\n%s", wanted, reasons)
	}
	readme, err := os.ReadFile(cfg.readmePath)
	if err != nil {
		t.Fatal(err)
	}
	for _, wanted := range []string{
		readmeSupportStart,
		"**2 of 3 cards eligible for paper support (66.7%)**",
		"[`supported.md`](./supported.md)",
		"[`unsupported-reasons.md`](./unsupported-reasons.md)",
		readmeSupportEnd,
	} {
		if !strings.Contains(string(readme), wanted) {
			t.Errorf("README missing %q:\n%s", wanted, readme)
		}
	}
}

func TestUpdateReadmeSupportRequiresMarkers(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "README.md")
	if err := os.WriteFile(path, []byte("# Missing markers\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := updateReadmeSupport(path, report{}); err == nil {
		t.Fatal("updateReadmeSupport() succeeded without markers")
	}
}

func TestCompileCorpusDisambiguatesPathAndIdentifierCollisions(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"one","oracle_id":"oracle-one","name":"Same Name","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature"},
		{"id":"two","oracle_id":"oracle-two","name":"Same Name","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature"}
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

func TestCompileCorpusCategorizesTokensByOracleIdentity(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"card","oracle_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa","name":"Bear","layout":"normal","games":["paper"],"legalities":{"commander":"legal"},"type_line":"Creature — Bear","power":"2","toughness":"2"},
		{"id":"token-one","oracle_id":"11111111-1111-1111-1111-111111111111","name":"Bear","layout":"token","set_type":"token","games":["paper"],"type_line":"Token Creature — Bear","power":"2","toughness":"2"},
		{"id":"token-two","oracle_id":"22222222-2222-2222-2222-222222222222","name":"Bear","layout":"token","set_type":"token","games":["paper"],"type_line":"Token Creature — Bear","power":"1","toughness":"1"}
	]`
	results, err := compileCorpus(strings.NewReader(input), 2)
	if err != nil {
		t.Fatal(err)
	}
	wanted := map[string]string{
		filepath.Join("b", "bear.go"):                                            "var Bear =",
		filepath.Join("tokens", "b", "bear_11111111111111111111111111111111.go"): "var BearToken11111111111111111111111111111111 =",
		filepath.Join("tokens", "b", "bear_22222222222222222222222222222222.go"): "var BearToken22222222222222222222222222222222 =",
	}
	for _, result := range results {
		if result.err != nil || len(result.diagnostics) != 0 {
			t.Fatalf("result = %#v", result)
		}
		variable, ok := wanted[result.relative]
		if !ok {
			t.Fatalf("unexpected relative path %q", result.relative)
		}
		if !strings.Contains(result.source, "package b") || !strings.Contains(result.source, variable) {
			t.Fatalf("source for %s:\n%s", result.relative, result.source)
		}
		delete(wanted, result.relative)
	}
	if len(wanted) != 0 {
		t.Fatalf("missing generated paths: %v", wanted)
	}
}

func TestCompileCorpusRejectsInvalidTokenOracleID(t *testing.T) {
	t.Parallel()
	input := `[
		{"id":"token","oracle_id":"invalid","name":"Bear","layout":"token","set_type":"token","games":["paper"],"type_line":"Token Creature — Bear","power":"2","toughness":"2"}
	]`
	results, err := compileCorpus(strings.NewReader(input), 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || len(results[0].diagnostics) != 1 ||
		results[0].diagnostics[0].Summary != "invalid generated identity" {
		t.Fatalf("results = %#v", results)
	}
}

func TestWriteSupportedMigratesAndReconcilesTokenPaths(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ordinaryPath := filepath.Join(root, "b", "bear.go")
	obsoleteTokenPath := filepath.Join(root, "tokens", "b", "bear_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa.go")
	for _, path := range []string{ordinaryPath, obsoleteTokenPath} {
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("package b\n"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	input := `[
		{"id":"token","oracle_id":"11111111-1111-1111-1111-111111111111","name":"Bear","layout":"token","set_type":"token","games":["paper"],"type_line":"Token Creature — Bear","power":"2","toughness":"2"}
	]`
	results, err := compileCorpus(strings.NewReader(input), 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := writeSupported(root, results); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{ordinaryPath, obsoleteTokenPath} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("obsolete path %s still exists or stat failed: %v", path, err)
		}
	}
	finalPath := filepath.Join(
		root,
		"tokens",
		"b",
		"bear_11111111111111111111111111111111.go",
	)
	if _, err := os.Stat(finalPath); err != nil {
		t.Fatalf("generated token source missing: %v", err)
	}
	for _, path := range []string{
		filepath.Join(root, "tokens", "README.md"),
		filepath.Join(root, "tokens", "b", "README.md"),
		filepath.Join(root, "tokens", "b", "cards.go"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("token package file %s missing: %v", path, err)
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
