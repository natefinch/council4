package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IngaAndEsika is the card definition for Inga and Esika.
//
// Type: Legendary Creature — Human God
// Cost: {2}{G}{U}
//
// Oracle text:
//
//	Creatures you control have vigilance and "{T}: Add one mana of any color. Spend this mana only to cast a creature spell."
//	Whenever you cast a creature spell, if three or more mana from creatures was spent to cast it, draw a card.
var IngaAndEsika = newIngaAndEsika

func newIngaAndEsika() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Inga and Esika",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.God},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Creatures you control have vigilance and \"{T}: Add one mana of any color. Spend this mana only to cast a creature spell.\"",
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Vigilance},
						},
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.ManaAbility{
									Text: "{T}: Add one mana of any color. Spend this mana only to cast a creature spell.",
									AdditionalCosts: []cost.Additional{
										cost.Additional{
											Kind:        cost.AdditionalTap,
											CounterKind: counter.PlusOnePlusOne,
										},
									},
									Content: game.AbilityContent{
										Modes: []game.Mode{
											game.Mode{
												Sequence: []game.Instruction{
													game.Instruction{
														Primitive: game.Choose{
															Choice: game.ResolutionChoice{
																Kind:   game.ResolutionChoiceMana,
																Prompt: "Choose a color",
																Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
															},
															PublishChoice: game.ChoiceKey("oracle-mana-color"),
														},
													},
													game.Instruction{
														Primitive: game.AddMana{
															Amount:     game.Fixed(1),
															ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
															SpendRider: opt.Val(game.ManaSpendRider{
																Condition:   game.ManaSpendCastCreatureSpell,
																Restriction: game.ManaSpendRestrictedToCondition,
															}),
														},
													},
												},
											},
										},
										MinModes: 1,
										MaxModes: 1,
									},
								},
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you cast a creature spell, if three or more mana from creatures was spent to cast it, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventSpellCast,
							Controller:      game.TriggerControllerYou,
							CardSelection:   game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if three or more mana from creatures was spent to cast it",
						InterveningCondition: opt.Val(game.Condition{
							Text: "if three or more mana from creatures was spent to cast it",
							Aggregates: []game.AggregateComparison{
								game.AggregateComparison{
									Aggregate: game.AggregateEventSpellManaFromCreaturesSpentToCast,
									Op:        compare.GreaterOrEqual,
									Value:     3,
								},
							},
							SourceCounterKind: counter.PlusOnePlusOne,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
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
			Creatures you control have vigilance and "{T}: Add one mana of any color. Spend this mana only to cast a creature spell."
			Whenever you cast a creature spell, if three or more mana from creatures was spent to cast it, draw a card.
		`,
		},
	}
}
