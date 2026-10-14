package preservation

import (
	"strings"
	"testing"
)

func testManifest() Manifest {
	return Manifest{
		SchemaVersion: 1,
		Input: Pin{
			DownloadURI: "https://example.test/frozen.jsonl.gz", DownloadSHA256: strings.Repeat("a", 64),
			JSONSHA256: strings.Repeat("b", 64), UpdatedAt: "2026-10-13",
			RulesEdition: "2026-09-25", BaselineRevision: "baseline",
		},
		Compiler: []Record{{Key: "oracle-id", Name: "Card", DefinitionSHA256: strings.Repeat("c", 64)}},
		Registry: []Record{{Key: "Card#0", Name: "Card", DefinitionSHA256: strings.Repeat("d", 64)}},
	}
}

func TestCompareLossesAndChanges(t *testing.T) {
	t.Parallel()
	for _, scenario := range []struct {
		name         string
		change       func(*Manifest)
		registryOnly bool
		exception    bool
		wantError    bool
		wantKind     string
	}{
		{name: "unchanged"},
		{name: "lost compiler card", change: func(m *Manifest) { m.Compiler = nil }, wantError: true, wantKind: "loss"},
		{name: "lost registry card", change: func(m *Manifest) { m.Registry = nil }, wantError: true, wantKind: "loss"},
		{name: "refused card", change: func(m *Manifest) {
			m.Compiler[0].DefinitionSHA256 = ""
			m.Compiler[0].Refusals = []string{"explicit refusal"}
		}, wantError: true, wantKind: "loss"},
		{name: "acknowledged loss", change: func(m *Manifest) { m.Compiler = nil }, exception: true, wantKind: "loss"},
		{name: "definition change", change: func(m *Manifest) { m.Compiler[0].DefinitionSHA256 = strings.Repeat("e", 64) }, wantKind: "changed"},
		{name: "registry-only does not recompile", change: func(m *Manifest) { m.Compiler = nil }, registryOnly: true},
		{name: "different corpus", change: func(m *Manifest) { m.Input.JSONSHA256 = strings.Repeat("f", 64) }, wantError: true},
		{name: "duplicate identity", change: func(m *Manifest) { m.Compiler = append(m.Compiler, m.Compiler[0]) }, wantError: true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			t.Parallel()
			base, candidate := testManifest(), testManifest()
			if scenario.change != nil {
				scenario.change(&candidate)
			}
			var exceptions []Exception
			if scenario.exception {
				exceptions = []Exception{{
					Inventory: "compiler", Key: "oracle-id", CorpusSHA256: base.Input.JSONSHA256,
					BaselineFingerprint: base.Compiler[0].DefinitionSHA256,
					Reason:              "approved correctness correction", IssueURL: "https://github.com/natefinch/council4/issues/3211",
				}}
			}
			changes, err := Compare(base, candidate, exceptions, scenario.registryOnly)
			if (err != nil) != scenario.wantError {
				t.Fatalf("error = %v, want error %v", err, scenario.wantError)
			}
			if scenario.wantKind != "" && (len(changes) != 1 || changes[0].Kind != scenario.wantKind) {
				t.Fatalf("changes = %+v, want %s", changes, scenario.wantKind)
			}
			if scenario.exception && !changes[0].Acknowledged {
				t.Fatal("loss was not acknowledged")
			}
		})
	}
}

func TestRejectStaleOrIncompleteExceptions(t *testing.T) {
	t.Parallel()
	base, candidate := testManifest(), testManifest()
	candidate.Compiler = nil
	for _, exception := range []Exception{
		{Inventory: "compiler", Key: "oracle-id"},
		{Inventory: "compiler", Key: "oracle-id", Reason: "reason", IssueURL: "https://example.test/issue",
			CorpusSHA256: base.Input.JSONSHA256, BaselineFingerprint: base.Compiler[0].DefinitionSHA256},
	} {
		if _, err := Compare(base, candidate, []Exception{exception}, false); err == nil {
			t.Fatal("accepted an incomplete loss exception")
		}
	}
}
