package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheSwarmweaver is the card definition for The Swarmweaver.
//
// Type: Legendary Artifact Creature — Scarecrow
// Cost: {2}{B}{G}
//
// Oracle text:
//
//	When The Swarmweaver enters, create two 1/1 black and green Insect creature tokens with flying.
//	Delirium — As long as there are four or more card types among cards in your graveyard, Insects and Spiders you control get +1/+1 and have deathtouch.
var TheSwarmweaver = newTheSwarmweaver

func newTheSwarmweaver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name: "The Swarmweaver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.G,
			}),
			Colors:     []color.Color{color.Black, color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact, types.Creature},
			Subtypes:   []types.Sub{types.Scarecrow},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Delirium — As long as there are four or more card types among cards in your graveyard, Insects and Spiders you control get +1/+1 and have deathtouch.",
					Condition: opt.Val(game.Condition{
						Text: "As long as there are four or more card types among cards in your graveyard",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerGraveyardCardTypeCount,
								Op:        compare.GreaterOrEqual,
								Value:     4,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Insect"), types.Sub("Spider")}}),
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Insect"), types.Sub("Spider")}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Deathtouch},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When The Swarmweaver enters, create two 1/1 black and green Insect creature tokens with flying.",
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
										Primitive: game.CreateToken{
											Amount: game.Fixed(2),
											Source: game.TokenDef(theSwarmweaverToken),
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
			When The Swarmweaver enters, create two 1/1 black and green Insect creature tokens with flying.
			Delirium — As long as there are four or more card types among cards in your graveyard, Insects and Spiders you control get +1/+1 and have deathtouch.
		`,
		},
	}
}

var theSwarmweaverToken = newTheSwarmweaverToken()

func newTheSwarmweaverToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Insect",
			Colors:    []color.Color{color.Black, color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Insect},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
		},
	}
}
