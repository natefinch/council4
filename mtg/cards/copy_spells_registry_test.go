package cards_test

import (
	"slices"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
)

func TestCopySpellsProductionIdentityAndVerification(t *testing.T) {
	t.Parallel()
	registry := cards.NewDefaultRegistry()
	for _, entry := range []cardset.Entry{authored.Twincast, authored.Reverberate, authored.DualcasterMage, authored.CacklingCounterpart, authored.FatedInfatuation} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name || len(registry.LookupAll(entry.Name)) != 1 {
				t.Fatalf("production identity %s: %v", entry.OracleID, err)
			}
			if err := registry.Programs().ValidateCard(def); err != nil {
				t.Fatal(err)
			}
			index := slices.IndexFunc(authored.Cards, func(candidate cardset.Entry) bool { return candidate.OracleID == entry.OracleID })
			if index < 0 {
				t.Fatal("verified entry is not registered")
			}
			verified := authored.Cards[index]
			if entry.Verification == "" || verified.Verification != entry.Verification {
				t.Fatal("specific scenario evidence was lost")
			}
		})
	}
	if registry.Lookup("Reiterate") != nil {
		t.Fatal("buyback-blocked Reiterate was partially promoted")
	}
}
