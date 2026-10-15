package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FullThrottle is the card definition for Full Throttle.
//
// Type: Sorcery
// Cost: {4}{R}{R}
//
// Oracle text:
//
//	After this main phase, there are two additional combat phases.
//	At the beginning of each combat this turn, untap all creatures that attacked this turn.
var FullThrottle = newFullThrottle

func newFullThrottle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Full Throttle",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddExtraPhases{
							CombatCount: 2,
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Untap{
														Group: game.AttackedThisTurnGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								EventPattern: opt.Val(game.TriggerPattern{
									Event:           game.EventBeginningOfStep,
									StackObjectKind: game.StackSpell,
									CounterKind:     counter.PlusOnePlusOne,
									Step:            game.StepBeginningOfCombat,
								}),
								Window: game.DelayedWindowThisTurn,
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			After this main phase, there are two additional combat phases.
			At the beginning of each combat this turn, untap all creatures that attacked this turn.
		`,
		},
	}
}
