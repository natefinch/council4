package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CleverImpersonator is the card definition for Clever Impersonator.
//
// Type: Creature — Shapeshifter
// Cost: {2}{U}{U}
//
// Oracle text:
//
//	You may have this creature enter as a copy of any nonland permanent on the battlefield.
var CleverImpersonator = newCleverImpersonator

func newCleverImpersonator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Clever Impersonator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any nonland permanent on the battlefield.", &game.Selection{ExcludedTypes: []types.Card{types.Land}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this creature enter as a copy of any nonland permanent on the battlefield.
		`,
		},
	}
}
