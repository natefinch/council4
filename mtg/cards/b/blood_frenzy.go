package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodFrenzy is the card definition for Blood Frenzy.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Cast this spell only before the combat damage step.
//	Target attacking or blocking creature gets +4/+0 until end of turn. Destroy that creature at the beginning of the next end step.
var BloodFrenzy = newBloodFrenzy

func newBloodFrenzy() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Blood Frenzy",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text:                           "Cast this spell only before the combat damage step.",
					CastOnlyBeforeCombatDamageStep: true,
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target attacking or blocking creature",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, CombatState: game.CombatStateAttackingOrBlocking}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.ModifyPT{
							Object:     game.TargetPermanentReference(0),
							PowerDelta: game.Fixed(4),
							Duration:   game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextEndStep,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.Destroy{
														Object: game.CapturedObjectReference(),
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
								CapturedObject: opt.Val(game.TargetPermanentReference(0)),
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Cast this spell only before the combat damage step.
			Target attacking or blocking creature gets +4/+0 until end of turn. Destroy that creature at the beginning of the next end step.
		`,
		},
	}
}
