package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RaugrinTriome is the card definition for Raugrin Triome.
//
// Type: Land — Island Mountain Plains
//
// Oracle text:
//
//	({T}: Add {U}, {R}, or {W}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var RaugrinTriome = newRaugrinTriome

func newRaugrinTriome() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Raugrin Triome",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Mountain, types.Plains},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U}, {R}, or {W}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
