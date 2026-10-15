package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JwariShapeshifter is the card definition for Jwari Shapeshifter.
//
// Type: Creature — Shapeshifter Ally
// Cost: {1}{U}
//
// Oracle text:
//
//	You may have this creature enter as a copy of any Ally creature on the battlefield.
var JwariShapeshifter = newJwariShapeshifter

func newJwariShapeshifter() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jwari Shapeshifter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter, types.Ally},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 0}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersAsCopyReplacement("You may have this creature enter as a copy of any Ally creature on the battlefield.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Ally")}}, true, false, nil, false, nil, nil),
			},
			OracleText: `
			You may have this creature enter as a copy of any Ally creature on the battlefield.
		`,
		},
	}
}
