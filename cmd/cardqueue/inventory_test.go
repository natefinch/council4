package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/natefinch/council4/cardgen"
	"github.com/natefinch/council4/internal/supportinventory"
	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/game"
)

func TestQueueInventoryWithoutHistoricalCompilerReport(t *testing.T) {
	t.Chdir(t.TempDir())
	q, corpus, _ := queueFixture(t)
	registry := cards.NewRegistry([]*game.CardDef{{CardFace: game.CardFace{Name: "existing"}}})
	if err := os.MkdirAll(filepath.Dir(compiledInstallationPath), 0o750); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "corpus.json")
	if err := os.WriteFile(path, []byte("not needed by retained-input path"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(compiledInstallationPath, []byte(`{"corpus_sha256":"pin","cards":[{"oracle_id":"a","name":"a","layout":"normal"}]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	inventory, err := loadQueueInventory(path, "missing-report.json", corpus, "pin", registry)
	if err != nil {
		t.Fatal(err)
	}
	if err := validate(q, corpus, inventory); err != nil {
		t.Fatal("retained inputs changed saved accounting", err)
	}
	a := inventory.Corpus[slices.IndexFunc(inventory.Corpus, func(r supportinventory.Record) bool { return r.OracleID == "a" })]
	b := inventory.Corpus[slices.IndexFunc(inventory.Corpus, func(r supportinventory.Record) bool { return r.OracleID == "b" })]
	if !a.CompilerAccepted || a.Production != nil || b.CompilerAccepted || len(inventory.Backlog) != 6 {
		t.Fatal("installation fabricated loadability, verification or compiler acceptance for uninstalled records")
	}
	for _, input := range []string{
		`{"corpus_sha256":"other","cards":[]}`,
		`{"corpus_sha256":"pin","cards":[{"oracle_id":"a","name":"wrong","layout":"normal"}]}`,
		`{"corpus_sha256":"pin","cards":[{"oracle_id":"missing","name":"missing","layout":"normal"}]}`,
		`{"corpus_sha256":"pin","cards":[{"oracle_id":"a","name":"a","layout":"normal"},{"oracle_id":"a","name":"a","layout":"normal"}]}`,
		`{"corpus_sha256":"pin","cards":[{"oracle_id":"excluded","name":"excluded","layout":"normal"}]}`,
		`{`,
	} {
		if err := os.WriteFile(compiledInstallationPath, []byte(input), 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := loadQueueInventory(path, "missing-report.json", corpus, "pin", registry); err == nil {
			t.Fatalf("untrusted/mismatched retained evidence accepted: %s", input)
		}
	}
}

func TestQueueInventoryHonorsExplicitFrozenReportAndRejectsMalformedReport(t *testing.T) {
	t.Chdir(t.TempDir())
	card := cardgen.ScryfallCard{
		OracleID: "one", ID: "printing-one", Name: "One", Layout: "normal",
		Games: []string{"paper"}, Legalities: map[string]string{"legacy": "legal"},
	}
	data, err := json.Marshal([]cardgen.ScryfallCard{card})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("corpus.json", data, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile("compiler.json", []byte(`{"eligible_count":1,"generated_count":1}`), 0o600); err != nil {
		t.Fatal(err)
	}
	inventory, err := loadQueueInventory("corpus.json", "compiler.json", []cardgen.ScryfallCard{card}, "pin", cards.NewRegistry())
	if err != nil || len(inventory.Corpus) != 1 || !inventory.Corpus[0].CompilerAccepted {
		t.Fatal("existing frozen-report behavior lost", err)
	}
	if err := os.WriteFile("compiler.json", []byte(`{`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadQueueInventory("corpus.json", "compiler.json", []cardgen.ScryfallCard{card}, "pin", cards.NewRegistry()); err == nil {
		t.Fatal("malformed supplied report silently fell back")
	}
}

func TestRetainedInventoryPreservesExactAndLegacyIdentityBoundaries(t *testing.T) {
	paper := func(id, name, layout string) cardgen.ScryfallCard {
		return cardgen.ScryfallCard{
			OracleID: id, ID: "printing-" + id, Name: name, Layout: layout, TypeLine: "Creature",
			Games: []string{"paper"}, Legalities: map[string]string{"legacy": "legal"},
		}
	}
	corpus := []cardgen.ScryfallCard{
		paper("spell", "Recall", "normal"), paper("prepared", "Scholar // Recall", "prepare"),
		paper("token", "Recall", "token"), paper("multiface", "Front // Back", "modal_dfc"),
		paper("authored", "Authored", "normal"), paper("unloadable", "Unloadable", "normal"),
	}
	corpus[1].CardFaces = []cardgen.ScryfallCardFace{{Name: "Scholar"}, {Name: "Recall"}}
	corpus[3].CardFaces = []cardgen.ScryfallCardFace{{Name: "Front"}, {Name: "Back"}}
	production := []cards.Support{
		{Key: "legacy-name:Recall/0", Name: "Recall", Loadable: true, Faces: []string{"Recall"}},
		{Key: "legacy-name:Front/0", Name: "Front", Loadable: true, Faces: []string{"Front", "Back"}},
		{Key: "authored/front", OracleID: "authored", Name: "Authored", Loadable: true, BehaviorVerified: true, Verification: "actual scenario"},
		{Key: "unloadable/front", OracleID: "unloadable", Name: "Unloadable", Loadable: false, Verification: "not loadable"},
	}
	inventory, err := inventoryFromInstallation(corpus, "pin", production, nil)
	if err != nil {
		t.Fatal(err)
	}
	var backlog []string
	for _, r := range inventory.Backlog {
		backlog = append(backlog, r.OracleID)
	}
	if !slices.Equal(backlog, []string{"prepared", "token", "unloadable"}) {
		t.Fatalf("same-name token, partial alternate face or unloadable authored support escaped backlog: %v", backlog)
	}
	if len(inventory.Corpus[1].LegacyCandidates) != 1 || inventory.Corpus[1].Production != nil ||
		inventory.Corpus[4].Production == nil || !inventory.Corpus[4].Production.BehaviorVerified {
		t.Fatal("lost exact evidence or invented exact legacy associations")
	}
}
