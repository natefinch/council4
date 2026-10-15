package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThoughtReflection is the card definition for Thought Reflection.
//
// Type: Enchantment
// Cost: {4}{U}{U}{U}
//
// Oracle text:
//
//	If you would draw a card, draw two cards instead.
var ThoughtReflection = newThoughtReflection

func newThoughtReflection() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Thought Reflection",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawCardMultiplierReplacement("If you would draw a card, draw two cards instead.", 2, false),
			},
			OracleText: `
			If you would draw a card, draw two cards instead.
		`,
		},
	}
}
