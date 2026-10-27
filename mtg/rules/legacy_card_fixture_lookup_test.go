package rules

import (
	"encoding/json"
	"testing"

	"github.com/natefinch/council4/internal/carddata"
	"github.com/natefinch/council4/mtg/game"
)

func legacyCard(t *testing.T, card carddata.ScryfallCard) *game.CardDef {
	t.Helper()
	return legacyCardDefs(t, card)[0]
}

// legacyCardDefs selects frozen definitions; Oracle text is an exact fixture
// identifier, never interpreted. New scenarios should construct typed cards.
func legacyCardDefs(t *testing.T, card carddata.ScryfallCard) []*game.CardDef {
	t.Helper()
	key, err := json.Marshal(card)
	if err != nil {
		t.Fatal(err)
	}
	build, ok := legacyCardFixtures[string(key)]
	if !ok {
		t.Fatalf("no frozen fixture for %s; construct a typed CardDef for new scenarios", card.Name)
	}
	defs := build()
	if len(defs) != 1 {
		t.Fatalf("fixture %s has %d definitions, want 1", card.Name, len(defs))
	}
	return defs
}
