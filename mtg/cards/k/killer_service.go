package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KillerService is the card definition for Killer Service.
//
// Type: Enchantment
//
// Oracle text:
//
//	When this enchantment enters, create a number of Food tokens equal to the number of opponents you have. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	At the beginning of your end step, you may pay {2} and sacrifice a token. If you do, create a 4/4 green Rhino Warrior creature token.
var KillerService = newKillerService()

func newKillerService() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Killer Service",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
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
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountOpponentCount,
										Multiplier: 1,
									}),
									Source: game.TokenDef(killerServiceToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {2} and sacrifice a token?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(2),
										}),
										AdditionalCosts: []cost.Additional{
											{
												Kind:         cost.AdditionalSacrifice,
												Text:         "sacrifice a token",
												Amount:       1,
												RequireToken: true,
											},
										},
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(killerServiceToken2),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this enchantment enters, create a number of Food tokens equal to the number of opponents you have. (They're artifacts with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			At the beginning of your end step, you may pay {2} and sacrifice a token. If you do, create a 4/4 green Rhino Warrior creature token.
		`,
		},
	}
}

var killerServiceToken = newKillerServiceToken()

func newKillerServiceToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}

var killerServiceToken2 = newKillerServiceToken2()

func newKillerServiceToken2() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rhino Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rhino, types.Warrior},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
