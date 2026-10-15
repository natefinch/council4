package supportinventory

import (
	"encoding/json"
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
