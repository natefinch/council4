package x

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Xorn is the card definition for Xorn.
//
// Type: Creature — Elemental
// Cost: {2}{R}
//
// Oracle text:
//
//	If you would create one or more Treasure tokens, instead create those tokens plus an additional Treasure token.
var Xorn = newXorn

func newXorn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Xorn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacementFiltered("If you would create one or more Treasure tokens, instead create those tokens plus an additional Treasure token.", &game.TokenCreationReplacementSpec{Multiplier: 1, Addend: 1, Subtypes: []types.Sub{types.Sub("Treasure")}, Filter: game.TriggerControllerYou}),
			},
			OracleText: `
			If you would create one or more Treasure tokens, instead create those tokens plus an additional Treasure token.
		`,
		},
	}
}
