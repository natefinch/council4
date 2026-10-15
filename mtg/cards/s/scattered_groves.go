package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ScatteredGroves is the card definition for Scattered Groves.
//
// Type: Land — Forest Plains
//
// Oracle text:
//
//	({T}: Add {G} or {W}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var ScatteredGroves = newScatteredGroves

func newScatteredGroves() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:     "Scattered Groves",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Plains},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G} or {W}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
