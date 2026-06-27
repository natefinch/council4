package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Wizard
//
// Type: Token Creature — Wizard
//
// Oracle text:
//   {1}, Sacrifice this creature: Counter target noncreature spell unless its controller pays {1}.

// WizardToken7fb12be3b0e84c5b8258076ed8c2aa77 is the card definition for Wizard.
var WizardToken7fb12be3b0e84c5b8258076ed8c2aa77 = newWizardToken7fb12be3b0e84c5b8258076ed8c2aa77()

func newWizardToken7fb12be3b0e84c5b8258076ed8c2aa77() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Wizard",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Sacrifice this creature: Counter target noncreature spell unless its controller pays {1}.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
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
			{1}, Sacrifice this creature: Counter target noncreature spell unless its controller pays {1}.
		`,
		},
	}
}
