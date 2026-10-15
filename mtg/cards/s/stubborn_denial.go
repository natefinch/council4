package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StubbornDenial is the card definition for Stubborn Denial.
//
// Type: Instant
// Cost: {U}
//
// Oracle text:
//
//	Counter target noncreature spell unless its controller pays {1}.
//	Ferocious — If you control a creature with power 4 or greater, counter that spell instead.
var StubbornDenial = newStubbornDenial

func newStubbornDenial() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Stubborn Denial",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target noncreature spell",
						Allow:      game.TargetAllowStackObject,
						Predicate: game.TargetPredicate{
							ExcludedSpellCardTypes: []types.Card{types.Creature},
							StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
						},
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {1}?",
								Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								ManaCost: opt.Val(cost.Mana{
									cost.O(1),
								}),
							},
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate: true,
								ControlsMatching: opt.Val(game.SelectionCount{
									Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Power: opt.Val(compare.Int{Op: compare.GreaterOrEqual, Value: 4})},
								}),
							}),
						}),
						PublishCondition: "condition-1",
						PublishResult:    game.ResultKey("unless-paid"),
					},
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
						ConditionGate: "condition-1",
						ResultGate: opt.Val(game.InstructionResultGate{
							Key:       "unless-paid",
							Succeeded: game.TriFalse,
						}),
					},
					{
						Primitive: game.CounterObject{
							Object: game.TargetStackObjectReference(0),
						},
						ConditionGate:       "condition-1",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Counter target noncreature spell unless its controller pays {1}.
			Ferocious — If you control a creature with power 4 or greater, counter that spell instead.
		`,
		},
	}
}
