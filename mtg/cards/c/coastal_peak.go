package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// CoastalPeak is the card definition for Coastal Peak.
//
// Type: Land — Island Mountain
//
// Oracle text:
//
//	({T}: Add {U} or {R}.)
//	This land enters tapped.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var CoastalPeak = newCoastalPeak

func newCoastalPeak() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Coastal Peak",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Mountain},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U} or {R}.)
			This land enters tapped.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
