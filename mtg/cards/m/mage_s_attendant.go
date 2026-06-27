package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MageSAttendant is the card definition for Mage's Attendant.
//
// Type: Creature — Cat Rogue
// Cost: {2}{W}
//
// Oracle text:
//
//	When this creature enters, create a 1/1 blue Wizard creature token with "{1}, Sacrifice this token: Counter target noncreature spell unless its controller pays {1}."
var MageSAttendant = newMageSAttendant()

func newMageSAttendant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Mage's Attendant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Rogue},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(mageSAttendantToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, create a 1/1 blue Wizard creature token with "{1}, Sacrifice this token: Counter target noncreature spell unless its controller pays {1}."
		`,
		},
	}
}

var mageSAttendantToken = newMageSAttendantToken()

func newMageSAttendantToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Wizard",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, Sacrifice this token: Counter target noncreature spell unless its controller pays {1}.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:   cost.AdditionalSacrificeSource,
							Text:   "Sacrifice this token",
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
		},
	}
}
