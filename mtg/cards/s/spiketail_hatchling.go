package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SpiketailHatchling is the card definition for Spiketail Hatchling.
//
// Type: Creature — Drake
// Cost: {1}{U}
//
// Oracle text:
//
//	Flying (This creature can't be blocked except by creatures with flying or reach.)
//	Sacrifice this creature: Counter target spell unless its controller pays {1}.
var SpiketailHatchling = newSpiketailHatchling()

func newSpiketailHatchling() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Spiketail Hatchling",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Drake},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Sacrifice this creature: Counter target spell unless its controller pays {1}.",
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this creature",
							Amount: 1,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
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
										Prompt: "Pay {1}?",
										Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
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
					}.Ability(),
				},
			},
			OracleText: `
			Flying (This creature can't be blocked except by creatures with flying or reach.)
			Sacrifice this creature: Counter target spell unless its controller pays {1}.
		`,
		},
	}
}
