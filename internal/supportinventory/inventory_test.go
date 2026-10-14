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
