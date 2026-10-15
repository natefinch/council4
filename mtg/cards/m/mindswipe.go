package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Mindswipe is the card definition for Mindswipe.
//
// Type: Instant
// Cost: {X}{U}{R}
//
// Oracle text:
//
//	Counter target spell unless its controller pays {X}. Mindswipe deals X damage to that spell's controller.
var Mindswipe = newMindswipe

func newMindswipe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name: "Mindswipe",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.R,
			}),
			Colors: []color.Color{color.Red, color.Blue},
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
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {X}?",
								Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								DynamicGenericManaCost: opt.Val(&game.DynamicAmount{
									Kind:        game.DynamicAmountX,
									CounterKind: counter.PlusOnePlusOne,
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
					{
						Primitive: game.Damage{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind: game.DynamicAmountX,
							}),
							Recipient: game.PlayerDamageRecipient(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Counter target spell unless its controller pays {X}. Mindswipe deals X damage to that spell's controller.
		`,
		},
	}
}
