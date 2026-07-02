package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VedalkenHumiliator is the card definition for Vedalken Humiliator.
//
// Type: Creature — Vedalken Wizard
// Cost: {3}{U}
//
// Oracle text:
//
//	Metalcraft — Whenever this creature attacks, if you control three or more artifacts, creatures your opponents control lose all abilities and have base power and toughness 1/1 until end of turn.
var VedalkenHumiliator = newVedalkenHumiliator()

func newVedalkenHumiliator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Vedalken Humiliator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Wizard},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
						InterveningIf: "if you control three or more artifacts",
						InterveningCondition: opt.Val(game.Condition{
							ControlsMatching: opt.Val(game.SelectionCount{
								Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								MinCount:  3,
							}),
						}),
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer:        game.LayerPowerToughnessSet,
											Group:        game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
											SetPower:     opt.Val(game.PT{Value: 1}),
											SetToughness: opt.Val(game.PT{Value: 1}),
										},
										game.ContinuousEffect{
											Layer:              game.LayerAbility,
											Group:              game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
											RemoveAllAbilities: true,
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Metalcraft — Whenever this creature attacks, if you control three or more artifacts, creatures your opponents control lose all abilities and have base power and toughness 1/1 until end of turn.
		`,
		},
	}
}
