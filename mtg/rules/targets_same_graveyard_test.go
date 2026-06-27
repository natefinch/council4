package rules

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
)

// sameGraveyardSpec is the multi-target card spec the "Exile up to N target cards
// from a single graveyard" family lowers to: an any-controller graveyard card
// target whose chosen set must all share one graveyard.
func sameGraveyardSpec(maxTargets int) game.TargetSpec {
	return game.TargetSpec{
		MinTargets:    0,
		MaxTargets:    maxTargets,
		Allow:         game.TargetAllowCard,
		TargetZone:    zone.Graveyard,
		SameGraveyard: true,
	}
}

// TestSameGraveyardSpecRejectsCrossGraveyardCombination proves the runtime never
// offers a multi-card target set drawn from two different graveyards when the
// spec carries SameGraveyard, while same-graveyard pairs and single-card choices
// remain legal. A card in a graveyard is in its owner's graveyard (CR 404.2), so
// "same graveyard" holds exactly when every card target shares one owner.
func TestSameGraveyardSpecRejectsCrossGraveyardCombination(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "P1 Card One",
		Types: []types.Card{types.Instant},
	}})
	addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "P1 Card Two",
		Types: []types.Card{types.Sorcery},
	}})
	addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "P2 Card One",
		Types: []types.Card{types.Instant},
	}})

	result := targetChoicesForSpecs(g, game.Player1, &game.CardDef{}, 0, []game.TargetSpec{sameGraveyardSpec(2)})
	if result.kind != targetLegalChoicesFound {
		t.Fatalf("result kind = %v, want targetLegalChoicesFound", result.kind)
	}

	var sawSameGraveyardPair, sawSingle bool
	for _, choice := range result.choices {
		owners := map[game.PlayerID]bool{}
		for _, target := range choice {
			card, ok := g.GetCardInstance(target.CardID)
			if !ok {
				t.Fatalf("target card %v not found", target.CardID)
			}
			owners[card.Owner] = true
		}
		if len(owners) > 1 {
			t.Fatalf("offered cross-graveyard target set %+v", choice)
		}
		if len(choice) == 2 {
			sawSameGraveyardPair = true
		}
		if len(choice) == 1 {
			sawSingle = true
		}
	}
	if !sawSameGraveyardPair {
		t.Fatal("expected a same-graveyard two-card choice to be offered")
	}
	if !sawSingle {
		t.Fatal("expected single-card choices to be offered")
	}
}

// TestSameGraveyardSpecValidatesChosenSet proves the chosen-set validation path
// (used when an action's targets are checked, not just enumerated) rejects a
// cross-graveyard pair and accepts a same-graveyard pair.
func TestSameGraveyardSpecValidatesChosenSet(t *testing.T) {
	g := game.NewGame([game.NumPlayers]game.PlayerConfig{})
	firstP1 := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "P1 Card One",
		Types: []types.Card{types.Instant},
	}})
	secondP1 := addCardToGraveyard(g, game.Player1, &game.CardDef{CardFace: game.CardFace{
		Name:  "P1 Card Two",
		Types: []types.Card{types.Sorcery},
	}})
	firstP2 := addCardToGraveyard(g, game.Player2, &game.CardDef{CardFace: game.CardFace{
		Name:  "P2 Card One",
		Types: []types.Card{types.Instant},
	}})
	spec := sameGraveyardSpec(2)

	same := []game.Target{currentCardTarget(t, g, firstP1), currentCardTarget(t, g, secondP1)}
	if !targetsSatisfySameGraveyard(g, &spec, same) {
		t.Fatal("same-graveyard pair was rejected, want accepted")
	}
	cross := []game.Target{currentCardTarget(t, g, firstP1), currentCardTarget(t, g, firstP2)}
	if targetsSatisfySameGraveyard(g, &spec, cross) {
		t.Fatal("cross-graveyard pair was accepted, want rejected")
	}
}
