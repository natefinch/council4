package cards_test

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
)

func TestRemovalTokenCardsProductionIdentityAndVerification(t *testing.T) {
	t.Parallel()
	registry := cards.NewDefaultRegistry()
	for _, entry := range []cardset.Entry{authored.GenerousGift, authored.Pongify, authored.RapidHybridization, authored.Resculpt} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name {
				t.Fatalf("production identity %s: %v", entry.OracleID, err)
			}
			if len(registry.LookupAll(entry.Name)) != 1 || len(def.ProgramReferences()) != 1 {
				t.Fatal("duplicate name or missing program")
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(authored.Cards, func(candidate cardset.Entry) bool { return candidate.OracleID == entry.OracleID })
			if index < 0 {
				t.Fatal("verified entry is not registered")
			}
			verified := authored.Cards[index]
			if entry.Verification == "" || verified.Verification != entry.Verification ||
				verified.SourceRevision == "" || verified.RulesRevision == "" {
				t.Fatal("explicit verification was lost or source/rules bindings were omitted")
			}
		})
	}
}
