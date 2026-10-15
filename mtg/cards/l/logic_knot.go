package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LogicKnot is the card definition for Logic Knot.
//
// Type: Instant
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
//	Counter target spell unless its controller pays {X}.
var LogicKnot = newLogicKnot

func newLogicKnot() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Logic Knot",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.DelveStaticBody,
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
				},
			}.Ability()),
			OracleText: `
			Delve (Each card you exile from your graveyard while casting this spell pays for {1}.)
			Counter target spell unless its controller pays {X}.
		`,
		},
	}
}
