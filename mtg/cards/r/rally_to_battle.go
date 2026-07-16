package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RallyToBattle is the card definition for Rally to Battle.
//
// Type: Instant
// Cost: {3}{W}
//
// Oracle text:
//
//	Creatures you control get +1/+3 until end of turn. Untap them.
var RallyToBattle = newRallyToBattle

func newRallyToBattle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Rally to Battle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:          game.LayerPowerToughnessModify,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									PowerDelta:     1,
									ToughnessDelta: 3,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Untap{
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Creatures you control get +1/+3 until end of turn. Untap them.
		`,
		},
	}
}
