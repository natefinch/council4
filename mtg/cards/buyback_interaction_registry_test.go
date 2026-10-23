package cards_test

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
)

func TestBuybackInteractionProductionIdentities(t *testing.T) {
	t.Parallel()
	registry := cards.NewDefaultRegistry()
	for _, entry := range []cardset.Entry{authored.Reiterate, authored.Allay, authored.Capsize} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name || len(registry.LookupAll(entry.Name)) != 1 {
				t.Fatalf("production identity %s: %v", entry.OracleID, err)
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatal(err)
			}
			if _, ok := def.BuybackKeyword(); !ok {
				t.Fatal("buyback declaration missing from the checked production definition")
			}
			if !slices.ContainsFunc(authored.Cards, func(candidate cardset.Entry) bool {
				return candidate.OracleID == entry.OracleID &&
					candidate.Name == entry.Name &&
					candidate.Verification == entry.Verification
			}) {
				t.Fatal("specific scenario evidence was lost from authored registration")
			}
		})
	}
}
