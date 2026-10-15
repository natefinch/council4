package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SeasonedPyromancer is the card definition for Seasoned Pyromancer.
//
// Type: Creature — Human Shaman
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	When this creature enters, discard two cards, then draw two cards. For each nonland card discarded this way, create a 1/1 red Elemental creature token.
//	{3}{R}{R}, Exile this card from your graveyard: Create two 1/1 red Elemental creature tokens.
var SeasonedPyromancer = newSeasonedPyromancer

func newSeasonedPyromancer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Seasoned Pyromancer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Shaman},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{3}{R}{R}, Exile this card from your graveyard: Create two 1/1 red Elemental creature tokens.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(3),
						cost.R,
						cost.R,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile this card from your graveyard",
							Amount:      1,
							Source:      zone.Graveyard,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Graveyard,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(seasonedPyromancerToken),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, discard two cards, then draw two cards. For each nonland card discarded this way, create a 1/1 red Elemental creature token.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
										PublishResult: game.ResultKey("actual-card-quantity-1"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("actual-card-quantity-1")},
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(2),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountPreviousEffectResult,
												Multiplier: 1,
												Selection:  &game.Selection{ExcludedTypes: []types.Card{types.Land}},
												ResultKey:  game.ResultKey("actual-card-quantity-1"),
											}),
											Source: game.TokenDef(seasonedPyromancerToken),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:      game.ResultKey("actual-card-quantity-1"),
											Accepted: game.TriTrue,
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
			When this creature enters, discard two cards, then draw two cards. For each nonland card discarded this way, create a 1/1 red Elemental creature token.
			{3}{R}{R}, Exile this card from your graveyard: Create two 1/1 red Elemental creature tokens.
		`,
		},
	}
}

var seasonedPyromancerToken = newSeasonedPyromancerToken()

func newSeasonedPyromancerToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Elemental",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
