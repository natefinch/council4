package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/opt"
)

func TestObservedCardPTRequiresExactIncarnationAndKnownCharacteristic(t *testing.T) {
	for _, tc := range []struct {
		name       string
		modify     func(*game.CardInstance)
		resolves   bool
		powerKnown bool
	}{
		{"known zero", func(card *game.CardInstance) { card.Def.Power = opt.Val(game.PT{}) }, true, true},
		{"undefined", func(card *game.CardInstance) { card.Def.Power = opt.V[game.PT]{} }, true, false},
		{"star", func(card *game.CardInstance) { card.Def.Power = opt.Val(game.PT{IsStar: true}) }, true, false},
		{"missing definition", func(card *game.CardInstance) { card.Def = nil }, false, false},
		{"same card new incarnation", func(card *game.CardInstance) { card.ZoneVersion++ }, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
			cardID := addCardToLibrary(g, game.Player1, vanillaCreature("Observed", 0, 3))
			card := g.CardInstances[cardID]
			ref := game.LinkedObjectRef{CardID: cardID, CardZoneVersion: card.ZoneVersion, CardZoneVersionSet: true}
			tc.modify(card)
			resolved, available := resolveLinkedObjectRef(g, ref)
			if available != tc.resolves {
				t.Fatalf("subject available=%t, want %t", available, tc.resolves)
			}
			if available && (resolved.snapshot.Power.Exists != tc.powerKnown ||
				resolved.snapshot.Power.Val != 0 || !resolved.snapshot.Toughness.Exists || resolved.snapshot.Toughness.Val != 3) {
				t.Fatalf("printed numeric availability = %#v", resolved.snapshot)
			}
			ref.ObjectID = g.IDGen.Next()
			if _, available := resolveLinkedObjectRef(g, ref); available {
				t.Fatal("missing nonzero object identity fell back to printed card numerics")
			}
		})
	}
}
