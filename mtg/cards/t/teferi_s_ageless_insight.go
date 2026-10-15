package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TeferiSAgelessInsight is the card definition for Teferi's Ageless Insight.
//
// Type: Legendary Enchantment
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.
var TeferiSAgelessInsight = newTeferiSAgelessInsight

func newTeferiSAgelessInsight() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Teferi's Ageless Insight",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DrawCardMultiplierReplacement("If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.", 2, true),
			},
			OracleText: `
			If you would draw a card except the first one you draw in each of your draw steps, draw two cards instead.
		`,
		},
	}
}
