package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WarriorSStand is the card definition for Warrior's Stand.
//
// Type: Instant
// Cost: {1}{W}
//
// Oracle text:
//
//	Cast this spell only during the declare attackers step and only if you've been attacked this step.
//	Creatures you control get +2/+2 until end of turn.
var WarriorSStand = newWarriorSStand

func newWarriorSStand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Warrior's Stand",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:                          "Cast this spell only during the declare attackers step and only if you've been attacked this step.",
					CastOnlyAfterAttackedThisStep: true,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Duration:       game.DurationPermanent,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									Layer:          game.LayerPowerToughnessModify,
									PowerDelta:     2,
									ToughnessDelta: 2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cast this spell only during the declare attackers step and only if you've been attacked this step.
			Creatures you control get +2/+2 until end of turn.
		`,
		},
	}
}
