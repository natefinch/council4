package cards_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
)

func TestSearchPartitionCardsProductionIdentityAndVerification(t *testing.T) {
	t.Parallel()
	registry := cards.NewDefaultRegistry()
	for _, entry := range []cardset.Entry{authored.Intuition, authored.RealmsUncharted} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name || def.OracleText != entry.New().OracleText {
				t.Fatalf("production Oracle identity %s: %v", entry.OracleID, err)
			}
			if len(registry.LookupAll(entry.Name)) != 1 || len(def.ProgramReferences()) != 1 {
				t.Fatal("duplicate name or missing authored program")
			}
			if issues := game.ValidateCardDef(def); len(issues) != 0 {
				t.Fatalf("invalid authored declarations: %+v", issues)
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(authored.Cards, func(candidate cardset.Entry) bool { return candidate.OracleID == entry.OracleID })
			if index < 0 {
				t.Fatal("complete authored card is not registered")
			}
			verified := authored.Cards[index]
			if !strings.Contains(entry.Verification, "mtg/rules/search_partition_spells_test.go") ||
				verified.Verification != entry.Verification {
				t.Fatal("specific scenario evidence was lost")
			}
		})
	}
}
