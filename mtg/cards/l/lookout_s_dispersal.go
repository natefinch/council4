package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LookoutSDispersal is the card definition for Lookout's Dispersal.
//
// Type: Instant
// Cost: {2}{U}
//
// Oracle text:
//
//	This spell costs {1} less to cast if you control a Pirate.
//	Counter target spell unless its controller pays {4}.
var LookoutSDispersal = newLookoutSDispersal

func newLookoutSDispersal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Lookout's Dispersal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								ReductionCondition: opt.Val(game.Condition{
									ControlsMatching: opt.Val(game.SelectionCount{
										Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Pirate")}},
									}),
								}),
							},
						},
					},
				},
			},
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
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {4}?",
								Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								ManaCost: opt.Val(cost.Mana{
									cost.O(4),
								}),
							},
						},
						PublishResult: game.ResultKey("unless-paid"),
					},
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "unless-paid",
							Succeeded: game.TriFalse,
						}),
					},
				},
			}.Ability()),
			OracleText: `
			This spell costs {1} less to cast if you control a Pirate.
			Counter target spell unless its controller pays {4}.
		`,
		},
	}
}
