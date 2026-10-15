package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SavaiTriome is the card definition for Savai Triome.
//
// Type: Land — Mountain Plains Swamp
//
// Oracle text:
//
//	({T}: Add {R}, {W}, or {B}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var SavaiTriome = newSavaiTriome

func newSavaiTriome() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Red),
		CardFace: game.CardFace{
			Name:     "Savai Triome",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Plains, types.Swamp},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {R}, {W}, or {B}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
