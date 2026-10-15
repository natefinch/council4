package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheCapitolineTriad is the card definition for The Capitoline Triad.
//
// Type: Legendary Creature — God Artificer
// Cost: {10}
//
// Oracle text:
//
//	Those Who Came Before — This spell costs {1} less to cast for each historic card in your graveyard. (Artifacts, legendaries, and Sagas are historic.)
//	Exile any number of historic cards from your graveyard with total mana value 30 or greater: You get an emblem with "Creatures you control have base power and toughness 9/9."
var TheCapitolineTriad = newTheCapitolineTriad

func newTheCapitolineTriad() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Capitoline Triad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(10),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.God, types.Artificer},
			Power:      opt.Val(game.PT{Value: 7}),
			Toughness:  opt.Val(game.PT{Value: 7}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Those Who Came Before — This spell costs {1} less to cast for each historic card in your graveyard. (Artifacts, legendaries, and Sagas are historic.)",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							Duration:       game.DurationPermanent,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{AnyOf: []game.Selection{game.Selection{RequiredTypes: []types.Card{types.Artifact}}, game.Selection{Supertypes: []types.Super{types.Legendary}}, game.Selection{SubtypesAny: []types.Sub{types.Saga}}}},
								CountZone:          opt.Val(zone.Graveyard),
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Exile any number of historic cards from your graveyard with total mana value 30 or greater: You get an emblem with \"Creatures you control have base power and toughness 9/9.\"",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:                  cost.AdditionalExile,
							Text:                  "Exile any number of historic cards from your graveyard with total mana value 30 or greater",
							MatchHistoric:         true,
							Source:                zone.Graveyard,
							CounterKind:           counter.PlusOnePlusOne,
							TotalManaValueAtLeast: 30,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateEmblem{
											EmblemAbilities: []game.Ability{
												&game.StaticAbility{
													Text: "Creatures you control have base power and toughness 9/9.",
													ContinuousEffects: []game.ContinuousEffect{
														game.ContinuousEffect{
															Duration: game.DurationPermanent,
															Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
															Layer:    game.LayerPowerToughnessSet,
															SetPower: opt.Val(game.PT{
																Value: 9,
															}),
															SetToughness: opt.Val(game.PT{
																Value: 9,
															}),
														},
													},
												},
											},
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
			Those Who Came Before — This spell costs {1} less to cast for each historic card in your graveyard. (Artifacts, legendaries, and Sagas are historic.)
			Exile any number of historic cards from your graveyard with total mana value 30 or greater: You get an emblem with "Creatures you control have base power and toughness 9/9."
		`,
		},
	}
}
