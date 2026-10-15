package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// KetriaTriome is the card definition for Ketria Triome.
//
// Type: Land — Forest Island Mountain
//
// Oracle text:
//
//	({T}: Add {G}, {U}, or {R}.)
//	This land enters tapped.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var KetriaTriome = newKetriaTriome

func newKetriaTriome() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Ketria Triome",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Island, types.Mountain},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G}, {U}, or {R}.)
			This land enters tapped.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
