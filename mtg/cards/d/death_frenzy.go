package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeathFrenzy is the card definition for Death Frenzy.
//
// Type: Sorcery
// Cost: {3}{B}{G}
//
// Oracle text:
//
//	All creatures get -2/-2 until end of turn. Whenever a creature dies this turn, you gain 1 life.
var DeathFrenzy = newDeathFrenzy

func newDeathFrenzy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "Death Frenzy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
				cost.G,
			}),
			Colors: []color.Color{color.Black, color.Green},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.ApplyContinuous{
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:          game.LayerPowerToughnessModify,
									Group:          game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									PowerDelta:     -2,
									ToughnessDelta: -2,
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								EventPattern: opt.Val(game.TriggerPattern{
									Event:            game.EventPermanentDied,
									SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
								}),
								Window: game.DelayedWindowThisTurn,
								Content: game.Mode{
									Sequence: []game.Instruction{
										{
											Primitive: game.GainLife{
												Amount: game.Fixed(1),
												Player: game.ControllerReference(),
											},
										},
									},
								}.Ability(),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			All creatures get -2/-2 until end of turn. Whenever a creature dies this turn, you gain 1 life.
		`,
		},
	}
}
