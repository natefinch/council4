package supportinventory

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
)

func TestIdentityAndProductionStatusDoNotFollowCompilerRejection(t *testing.T) {
	corpus := []OracleCard{{OracleID: "authored", Name: "Authored"}, {OracleID: "first", Name: "Ambiguous"}, {OracleID: "second", Name: "Ambiguous"}}
	production := []cards.Support{
		{Key: "authored/front", OracleID: "authored", Name: "Authored", Loadable: true, BehaviorVerified: true, Implementation: "authored"},
		{Key: "legacy-name:Ambiguous/0", Name: "Ambiguous", Identity: "name-only", Loadable: true},
	}
	report := CompilerReport{Eligible: 3, Generated: 2, Unsupported: []DiagnosticCard{
		{OracleID: "authored", Name: "Authored", Diagnostics: json.RawMessage(`[{"summary":"compiler refuses"}]`)},
	}}
	inventory, err := Merge(corpus, report, production, "corpus", "compiler")
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Production) != 2 || len(inventory.Backlog) != 0 {
		t.Fatal("lost preserved legacy entry or put a working authored identity in production backlog")
	}
	for _, record := range inventory.Corpus {
		if record.OracleID == "authored" {
			if record.CompilerAccepted || record.Production == nil || !record.Production.BehaviorVerified || len(record.Diagnostics) == 0 {
				t.Fatal("compiler refusal overrode working production behavior or diagnostics disappeared")
			}
		} else if record.Production != nil || len(record.LegacyCandidates) != 1 {
			t.Fatal("ambiguous legacy name was invented as exact Oracle identity")
		}
	}
}

func TestMismatchedCorpusEvidenceFails(t *testing.T) {
	if _, err := Merge([]OracleCard{{OracleID: "one", Name: "One"}}, CompilerReport{Eligible: 2, Generated: 2}, nil, "", ""); err == nil {
		t.Fatal("different compiler/corpus inputs were silently merged")
	}
}

func TestInstalledSpellDoesNotPromotePreparedCardOrSameNameToken(t *testing.T) {
	corpus := []OracleCard{
		{OracleID: "spell", Name: "Recall", Layout: "normal"},
		{OracleID: "prepared", Name: "Scholar // Recall", Layout: "prepare", Faces: []OracleFace{{Name: "Scholar"}, {Name: "Recall"}}},
		{OracleID: "token", Name: "Recall", Layout: "token"},
		{OracleID: "multiface", Name: "Front // Back", Layout: "modal_dfc", Faces: []OracleFace{{Name: "Front"}, {Name: "Back"}}},
	}
	production := []cards.Support{
		{Key: "legacy-name:Recall/0", Name: "Recall", Loadable: true, Faces: []string{"Recall"}},
		{Key: "legacy-name:Front/0", Name: "Front", Loadable: true, Faces: []string{"Front", "Back"}},
	}
	inventory, err := Merge(corpus, CompilerReport{Eligible: 4, Generated: 4}, production, "corpus", "compiler")
	if err != nil {
		t.Fatal(err)
	}
	if len(inventory.Backlog) != 2 {
		t.Fatalf("partial face or same-name token was mistaken for a loadable whole card: %+v", inventory.Backlog)
	}
	for _, record := range inventory.Backlog {
		if record.OracleID != "prepared" && record.OracleID != "token" {
			t.Fatalf("complete installed definition was lost: %+v", record)
		}
	}
}

func TestReportsRetainScenarioEvidenceWithoutSourceHashes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for name, data := range map[string]string{
		"README.md":              "# Project\n\n<!-- card-support:start -->\nold\n<!-- card-support:end -->\n",
		"card-backlog.md":        "# Historical backlog\n\nSaved diagnostics.\n",
		"unsupported-reasons.md": "# Historical reasons\n\nSaved diagnostics.\n",
	} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	support := cards.Support{
		Name: "Authored", OracleID: "authored", Loadable: true,
		Implementation: "authored", BehaviorVerified: true,
		Verification: "mtg/rules/example_test.go",
	}
	inventory := Inventory{
		Production: []cards.Support{support},
		Corpus:     []Record{{OracleID: support.OracleID, Name: support.Name, Production: &support}},
	}
	if err := inventory.WriteDocs(root); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(root, "support-inventory.md"))
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if !strings.Contains(text, support.Verification) ||
		!strings.Contains(text, ".github/copilot-instructions.md") ||
		!strings.Contains(text, "explicit reporting maintenance") {
		t.Fatal("report lost scenario evidence or the current validation policy")
	}
	if strings.Contains(text, "go generate") || strings.Contains(text, "sha256:") {
		t.Fatal("report retained executable-source fingerprint bookkeeping")
	}
}
