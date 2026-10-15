package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MartyrOfFrost is the card definition for Martyr of Frost.
//
// Type: Creature — Human Wizard
// Cost: {U}
//
// Oracle text:
//
//	{2}, Reveal X blue cards from your hand, Sacrifice this creature: Counter target spell unless its controller pays {X}.
var MartyrOfFrost = newMartyrOfFrost

func newMartyrOfFrost() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Martyr of Frost",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, Reveal X blue cards from your hand, Sacrifice this creature: Counter target spell unless its controller pays {X}.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:           cost.AdditionalReveal,
							Text:           "Reveal X blue cards from your hand",
							AmountFromX:    true,
							MatchCardColor: true,
							CardColor:      color.Blue,
							Source:         zone.Hand,
							CounterKind:    counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
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
									game.Instruction{
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
									game.Instruction{
										Primitive: game.CounterObject{
											Object: game.TargetStackObjectReference(0),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("unless-paid"),
											Succeeded: game.TriFalse,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			{2}, Reveal X blue cards from your hand, Sacrifice this creature: Counter target spell unless its controller pays {X}.
		`,
		},
	}
}
