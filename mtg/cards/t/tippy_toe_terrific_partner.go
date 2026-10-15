package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TippyToeTerrificPartner is the card definition for Tippy-Toe, Terrific Partner.
//
// Type: Legendary Creature — Squirrel Hero
// Cost: {3}{G}
//
// Oracle text:
//
//	If you would create one or more tokens, instead create those tokens plus an additional Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
//	At the beginning of your end step, if you gained life this turn, draw a card.
var TippyToeTerrificPartner = newTippyToeTerrificPartner

func newTippyToeTerrificPartner() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tippy-Toe, Terrific Partner",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Squirrel, types.Hero},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "At the beginning of your end step, if you gained life this turn, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepEnd,
						},
						InterveningIf: "if you gained life this turn",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if you gained life this turn",
							SourceCounterKind: counter.PlusOnePlusOne,
							EventHistory: opt.Val(game.EventHistoryCondition{
								Pattern: game.TriggerPattern{
									Event:           game.EventLifeGained,
									Player:          game.TriggerPlayerYou,
									StackObjectKind: game.StackSpell,
									CounterKind:     counter.PlusOnePlusOne,
								},
								Window: game.EventHistoryCurrentTurn,
							}),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.TokenCreationReplacementFiltered("If you would create one or more tokens, instead create those tokens plus an additional Food token. (It's an artifact with \"{2}, {T}, Sacrifice this token: You gain 3 life.\")", &game.TokenCreationReplacementSpec{Multiplier: 1, Addend: 1, AddendDef: tippyToeTerrificPartnerToken, Filter: game.TriggerControllerYou}),
			},
			OracleText: `
			If you would create one or more tokens, instead create those tokens plus an additional Food token. (It's an artifact with "{2}, {T}, Sacrifice this token: You gain 3 life.")
			At the beginning of your end step, if you gained life this turn, draw a card.
		`,
		},
	}
}

var tippyToeTerrificPartnerToken = newTippyToeTerrificPartnerToken()

func newTippyToeTerrificPartnerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Food",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Food},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice this artifact: You gain 3 life.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
							CounterKind:        counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GainLife{
											Amount: game.Fixed(3),
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
		},
	}
}
