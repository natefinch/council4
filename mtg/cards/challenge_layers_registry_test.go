package cards_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/natefinch/council4/mtg/cards"
	"github.com/natefinch/council4/mtg/cards/authored"
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
)

func TestChallengeLayersAndDrawProductionIdentityVerificationAndJSON(t *testing.T) {
	t.Parallel()
	registry := cards.NewDefaultRegistry()
	for _, entry := range []cardset.Entry{authored.ChainsOfMephistopheles, authored.Humility, authored.Opalescence} {
		t.Run(entry.Name, func(t *testing.T) {
			t.Parallel()
			def, err := registry.LookupOracleID(entry.OracleID)
			if err != nil || def == nil || def.Name != entry.Name || def.OracleText != entry.New().OracleText {
				t.Fatalf("production identity %s: %v", entry.OracleID, err)
			}
			if len(registry.LookupAll(entry.Name)) != 1 {
				t.Fatal("duplicate authored registration")
			}
			if issues := game.ValidateCardDef(def); len(issues) != 0 {
				t.Fatalf("invalid authored declaration: %+v", issues)
			}
			encoded, err := json.Marshal(def)
			if err != nil {
				t.Fatal(err)
			}
			var restored game.CardDef
			if err := json.Unmarshal(encoded, &restored); err != nil {
				t.Fatal(err)
			}
			if issues := game.ValidateCardDef(&restored); len(issues) != 0 {
				t.Fatalf("invalid serialized declaration: %+v", issues)
			}
			if err := registry.Programs().ValidateCard(&restored); err != nil {
				t.Fatal(err)
			}
			if entry.Name == "Opalescence" && (!strings.Contains(string(encoded), "\"SetPowerToManaValue\":true") ||
				!strings.Contains(string(encoded), "\"SetToughnessToManaValue\":true")) {
				t.Fatal("mana-value characteristic setters did not serialize")
			}
			index := slices.IndexFunc(authored.Cards, func(candidate cardset.Entry) bool { return candidate.OracleID == entry.OracleID })
			if index < 0 {
				t.Fatal("verified card not registered")
			}
			registered := authored.Cards[index]
			if !strings.Contains(entry.Verification, "mtg/rules/") ||
				registered.Verification != entry.Verification {
				t.Fatal("specific scenario evidence was lost")
			}
		})
	}
}
