package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Ultramarines Honour Guard
//
// Type: Token Creature — Astartes Warrior
//
// Oracle text:
//   Other creatures you control get +1/+1.

// UltramarinesHonourGuardToken00bc99d92cc44dd5a37b4918fff8160a is the card definition for Ultramarines Honour Guard.
var UltramarinesHonourGuardToken00bc99d92cc44dd5a37b4918fff8160a = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Ultramarines Honour Guard",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Astartes, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.StaticAbility{
				Text: "Other creatures you control get +1/+1.",
				ContinuousEffects: []game.ContinuousEffect{
					game.ContinuousEffect{
						Layer:          game.LayerPowerToughnessModify,
						Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}, game.SourcePermanentReference()),
						PowerDelta:     1,
						ToughnessDelta: 1,
					},
				},
			},
		},
		OracleText: `
			Other creatures you control get +1/+1.
		`,
	},
}
