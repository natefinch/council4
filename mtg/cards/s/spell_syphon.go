package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SpellSyphon is the card definition for Spell Syphon.
//
// Type: Instant
// Cost: {1}{U}
//
// Oracle text:
//
//	Counter target spell unless its controller pays {1} for each blue permanent you control.
var SpellSyphon = newSpellSyphon

func newSpellSyphon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Spell Syphon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
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
						Primitive: game.Pay{
							Payment: game.ResolutionPayment{
								Prompt: "Pay {1} for each blue permanent you control?",
								Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
								ManaCost: opt.Val(cost.Mana{
									cost.O(1),
								}),
								ManaCostMultiplier: opt.Val(&game.DynamicAmount{
									Kind:        game.DynamicAmountCountSelector,
									Multiplier:  1,
									CounterKind: counter.PlusOnePlusOne,
									Group:       game.BattlefieldGroup(game.Selection{ColorsAny: []color.Color{color.Blue}, Controller: game.ControllerYou}),
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
			Counter target spell unless its controller pays {1} for each blue permanent you control.
		`,
		},
	}
}
