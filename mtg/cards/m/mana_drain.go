package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ManaDrain is the card definition for Mana Drain.
//
// Type: Instant
// Cost: {U}{U}
//
// Oracle text:
//
//	Counter target spell. At the beginning of your next main phase, add an amount of {C} equal to that spell's mana value.
var ManaDrain = newManaDrain

func newManaDrain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Mana Drain",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							StackObjectKinds: []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
					},
					{
						Primitive: game.CreateDelayedTrigger{
							Trigger: game.DelayedTriggerDef{
								Timing: game.DelayedAtBeginningOfNextMainPhase,
								Content: game.AbilityContent{
									Modes: []game.Mode{
										game.Mode{
											Sequence: []game.Instruction{
												game.Instruction{
													Primitive: game.AddMana{
														Amount: game.Dynamic(game.DynamicAmount{
															Kind:   game.DynamicAmountCapturedTargetManaValue,
															Object: game.CapturedTargetStackObjectReference(0),
														}),
														ManaColor: mana.C,
													},
												},
											},
										},
									},
									MinModes: 1,
									MaxModes: 1,
								},
							},
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell. At the beginning of your next main phase, add an amount of {C} equal to that spell's mana value.
		`,
		},
	}
}
