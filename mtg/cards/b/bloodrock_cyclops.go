package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodrockCyclops is the card definition for Bloodrock Cyclops.
//
// Type: Creature — Cyclops
// Cost: {2}{R}
//
// Oracle text:
//
//	This creature attacks each combat if able.
var BloodrockCyclops = newBloodrockCyclops

func newBloodrockCyclops() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Bloodrock Cyclops",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cyclops},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.MustAttackStaticBody,
			},
			OracleText: `
			This creature attacks each combat if able.
		`,
		},
	}
}
