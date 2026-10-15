package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MyojinOfLifeSWeb is the card definition for Myojin of Life's Web.
//
// Type: Legendary Creature — Spirit
// Cost: {6}{G}{G}{G}
//
// Oracle text:
//
//	Myojin of Life's Web enters with a divinity counter on it if you cast it from your hand.
//	Myojin of Life's Web has indestructible as long as it has a divinity counter on it.
//	Remove a divinity counter from Myojin of Life's Web: Put any number of creature cards from your hand onto the battlefield.
var MyojinOfLifeSWeb = newMyojinOfLifeSWeb

func newMyojinOfLifeSWeb() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Myojin of Life's Web",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 8}),
			Toughness:  opt.Val(game.PT{Value: 8}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Myojin of Life's Web has indestructible as long as it has a divinity counter on it.",
					Condition: opt.Val(game.Condition{
						Text:   "as long as it has a divinity counter on it",
						Object: opt.Val(game.SourcePermanentReference()),
						ObjectMatches: opt.Val(game.Selection{RequiredCounter: counter.Divinity, RequiredCounterCount: opt.Val(compare.Int{
							Op:    compare.GreaterOrEqual,
							Value: 1,
						})}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Indestructible},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Remove a divinity counter from Myojin of Life's Web: Put any number of creature cards from your hand onto the battlefield.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a divinity counter from Myojin of Life's Web",
							Amount:      1,
							CounterKind: counter.Divinity,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
											Count:      game.ChooseAnyNumber,
											Destination: game.ChooseDestination{
												Zone: zone.Battlefield,
											},
											Prompt: "Choose a card to put onto the battlefield",
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
				game.EntersWithCountersIfReplacement("Myojin of Life's Web enters with a divinity counter on it if you cast it from your hand.", &game.Condition{
					EventPermanentWasCastFromControllerHand: true,
				}, game.CounterPlacement{Kind: counter.Divinity, Amount: 1}),
			},
			OracleText: `
			Myojin of Life's Web enters with a divinity counter on it if you cast it from your hand.
			Myojin of Life's Web has indestructible as long as it has a divinity counter on it.
			Remove a divinity counter from Myojin of Life's Web: Put any number of creature cards from your hand onto the battlefield.
		`,
		},
	}
}
