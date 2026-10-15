package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RagingRedcap is the card definition for Raging Redcap.
//
// Type: Creature — Goblin Knight
// Cost: {2}{R}
//
// Oracle text:
//
//	Double strike (This creature deals both first-strike and regular combat damage.)
var RagingRedcap = newRagingRedcap

func newRagingRedcap() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Raging Redcap",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Knight},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.DoubleStrikeStaticBody,
			},
			OracleText: `
			Double strike (This creature deals both first-strike and regular combat damage.)
		`,
		},
	}
}
