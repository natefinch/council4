package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JorKadeenThePrevailer is the card definition for Jor Kadeen, the Prevailer.
//
// Type: Legendary Creature — Human Warrior
// Cost: {3}{R}{W}
//
// Oracle text:
//
//	First strike
//	Metalcraft — Creatures you control get +3/+0 as long as you control three or more artifacts.
var JorKadeenThePrevailer = newJorKadeenThePrevailer

func newJorKadeenThePrevailer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Jor Kadeen, the Prevailer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.W,
			}),
			Colors:     []color.Color{color.Red, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.StaticAbility{
					Text: "Metalcraft — Creatures you control get +3/+0 as long as you control three or more artifacts.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control three or more artifacts",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							MinCount:  3,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:   game.DurationPermanent,
							Group:      game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:      game.LayerPowerToughnessModify,
							PowerDelta: 3,
						},
					},
				},
			},
			OracleText: `
			First strike
			Metalcraft — Creatures you control get +3/+0 as long as you control three or more artifacts.
		`,
		},
	}
}
