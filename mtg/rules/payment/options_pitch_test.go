package payment

import (
	"testing"

	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

func TestPitchAlternativeOptionPreservesExactExileAndOtherCosts(t *testing.T) {
	t.Parallel()
	required := cost.Additional{Kind: cost.AdditionalPayLife, Amount: 1}
	pitch := cost.Additional{
		Kind:           cost.AdditionalExile,
		Amount:         2,
		Source:         zone.Hand,
		MatchCardColor: true,
		CardColor:      color.Blue,
	}
	card := &game.CardDef{CardFace: game.CardFace{
		Name:            "Pitch Test",
		ManaCost:        opt.Val(cost.Mana{cost.O(5), cost.U, cost.U}),
		Types:           []types.Card{types.Instant},
		AdditionalCosts: []cost.Additional{required},
		AlternativeCosts: []cost.Alternative{{
			Label:           "Exile 2 blue cards",
			AdditionalCosts: []cost.Additional{pitch},
		}},
	}}
	options := spellCostOptionsForZoneAndKicker(fakePaymentState{}, game.Player1, card, zone.Hand, false, 0, false, nil)
	alternative, ok := spellOptionByLabel(options, "Exile 2 blue cards")
	if !ok {
		t.Fatal("pitch option missing")
	}
	if alternative.manaCost != nil {
		t.Fatalf("pitch mana cost = %#v, want no mana payment", alternative.manaCost)
	}
	if len(alternative.additionalCosts) != 2 ||
		alternative.additionalCosts[0] != required ||
		alternative.additionalCosts[1] != pitch {
		t.Fatalf("additional costs = %#v, want required cost then exact pitch cost", alternative.additionalCosts)
	}
}
