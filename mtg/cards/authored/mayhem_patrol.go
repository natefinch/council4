package authored

import (
	"github.com/natefinch/council4/mtg/cards/cardset"
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

const (
	mayhemPatrolOracleID   = "f77a9160-1d37-476d-8908-29ddcc982328"
	mayhemPatrolAttackText = "Whenever this creature attacks, target creature gets +1/+0 until end of turn."
	mayhemPatrolOracleText = `Menace (This creature can't be blocked except by two or more creatures.)
Whenever this creature attacks, target creature gets +1/+0 until end of turn.
Blitz {1}{R} (If you cast this spell for its blitz cost, it gains haste and "When this creature dies, draw a card." Sacrifice it at the beginning of the next end step.)`
)

var MayhemPatrol = cardset.Entry{
	OracleID: mayhemPatrolOracleID,
	Name:     "Mayhem Patrol",
	New:      newMayhemPatrol,
	Verification: "mtg/rules/mayhem_patrol_test.go: " +
		"TestMayhemPatrolBlitzAttackTargetsAnyCreatureAndGivesPlusOnePower",
}

func newMayhemPatrol() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Mayhem Patrol",
			ManaCost:   opt.Val(cost.Mana{cost.O(1), cost.R}),
			Colors:     []color.Color{color.Red},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Devil, types.Warrior},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 2}),
			OracleText: mayhemPatrolOracleText,
			AlternativeCosts: []cost.Alternative{
				cost.BlitzAlternative(cost.Mana{cost.O(1), cost.R}),
			},
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.BlitzStaticAbility(),
			},
			TriggeredAbilities: []game.TriggeredAbility{{
				Text: mayhemPatrolAttackText,
				Trigger: game.TriggerCondition{
					Type: game.TriggerWhenever,
					Pattern: game.TriggerPattern{
						Event:  game.EventAttackerDeclared,
						Source: game.TriggerSourceSelf,
					},
				},
				Content: game.Mode{
					Targets: []game.TargetSpec{{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature",
						Allow:      game.TargetAllowPermanent,
						Selection: opt.Val(game.Selection{
							RequiredTypes: []types.Card{types.Creature},
						}),
					}},
					Sequence: []game.Instruction{{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{{
								Layer:      game.LayerPowerToughnessModify,
								PowerDelta: 1,
							}},
							Duration: game.DurationUntilEndOfTurn,
						},
					}},
				}.Ability(),
			}},
		},
	}
}
