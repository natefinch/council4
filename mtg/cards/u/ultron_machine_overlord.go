package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UltronMachineOverlord is the card definition for Ultron, Machine Overlord.
//
// Type: Legendary Artifact Creature — Robot Villain
// Cost: {5}
//
// Oracle text:
//
//	Flying
//	Other Robots and Constructs you control get +2/+2.
var UltronMachineOverlord = newUltronMachineOverlord

func newUltronMachineOverlord() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ultron, Machine Overlord",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Robot, types.Villain},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Other Robots and Constructs you control get +2/+2.",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroupExcluding(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Robot"), types.Sub("Construct")}}, game.SourcePermanentReference()),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			OracleText: `
			Flying
			Other Robots and Constructs you control get +2/+2.
		`,
		},
	}
}
