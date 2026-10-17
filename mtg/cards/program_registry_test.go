package cards_test

import (
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cardprogram"
	"github.com/natefinch/council4/mtg/cards"
	authoredcards "github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

func TestDefaultRegistryAdmitsCompletedChoiceProgramsByExactIdentity(t *testing.T) {
	t.Parallel()
	registry := cards.NewDefaultRegistry()
	for _, entry := range []cardset.Entry{authoredcards.HiveMind, authoredcards.GiftsUngiven, authoredcards.GildedDrake, authoredcards.Mindslaver} {
		def, err := registry.LookupOracleID(entry.OracleID)
		if err != nil || def == nil || def.Name != entry.Name {
			t.Fatalf("exact production identity %s: %v", entry.OracleID, err)
		}
		if len(registry.LookupAll(entry.Name)) != 1 || len(def.ProgramReferences()) != 1 {
			t.Fatal("authored admission duplicated a name or lost its binding")
		}
		if err := registry.Programs().ValidateCard(def); err != nil {
			t.Fatal(err)
		}
	}
}

func TestAuthoredRegistryLoadsOnlyExactInstalledPrograms(t *testing.T) {
	t.Parallel()
	ref := game.ProgramReference{ID: "test/card", Version: "1"}
	entry := cardset.Entry{
		Name: "Authored", OracleID: "oracle-identity",
		Verification: "mtg/cards/program_registry_test.go: TestAuthoredRegistryLoadsOnlyExactInstalledPrograms (test fixture)",
		Programs:     []cardprogram.Program{{Reference: ref, Resolve: func(cardprogram.Context) error { return nil }}},
		New: func() *game.CardDef {
			return &game.CardDef{CardFace: game.CardFace{Name: "Authored", Types: []types.Card{types.Sorcery},
				SpellAbility: opt.Val(game.Mode{Program: ref}.Ability())}}
		},
	}
	registry, err := cards.NewProgramRegistry([]cardset.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}
	def, err := registry.LookupOracleID(entry.OracleID)
	if err != nil || def == nil {
		t.Fatalf("exact lookup: %v", err)
	}
	if _, err := registry.Programs().Lookup(ref); err != nil {
		t.Fatal(err)
	}
	entry.Programs = nil
	registry, err = cards.NewProgramRegistry([]cardset.Entry{entry})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := registry.LookupChecked(entry.Name); err == nil || !strings.Contains(err.Error(), ref.ID) {
		t.Fatalf("missing program diagnostic = %v", err)
	}
}

func TestOracleCollisionsAndExplicitLegacyReplacement(t *testing.T) {
	t.Parallel()
	construct := func() *game.CardDef { return &game.CardDef{CardFace: game.CardFace{Name: "Card"}} }
	legacy := cardset.Entry{Name: "Card", New: construct}
	authored := cardset.Entry{
		Name: "Card", OracleID: "oracle-identity", New: construct,
		Verification: "mtg/cards/program_registry_test.go: TestOracleCollisionsAndExplicitLegacyReplacement (test fixture)",
	}
	for _, entries := range [][]cardset.Entry{{legacy, authored}, {authored, legacy}} {
		if _, err := cards.NewProgramRegistry(entries); err == nil || !strings.Contains(err.Error(), "collides with an existing name") {
			t.Fatalf("name collision diagnostic = %v", err)
		}
	}
	authored.ReplaceLegacy = true
	registry, err := cards.NewProgramRegistry([]cardset.Entry{legacy, authored})
	if err != nil || registry.Len() != 1 {
		t.Fatalf("explicit replacement = %v", err)
	}
	if registry, err := cards.NewProgramRegistry([]cardset.Entry{authored, legacy}); err != nil || registry.Len() != 1 {
		t.Fatalf("reverse explicit replacement = %v", err)
	}
	if _, err := cards.NewProgramRegistry([]cardset.Entry{legacy, legacy, authored}); err == nil || !strings.Contains(err.Error(), "explicit unique legacy replacement required") {
		t.Fatalf("ambiguous legacy replacement diagnostic = %v", err)
	}
	authored.ReplaceLegacy = false
	if _, err := cards.NewProgramRegistry([]cardset.Entry{authored, authored}); err == nil || !strings.Contains(err.Error(), "duplicate Oracle") {
		t.Fatal("duplicate Oracle identity was accepted")
	}
}
