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

// MyojinOfSeeingWinds is the card definition for Myojin of Seeing Winds.
//
// Type: Legendary Creature — Spirit
// Cost: {7}{U}{U}{U}
//
// Oracle text:
//
//	Myojin of Seeing Winds enters with a divinity counter on it if you cast it from your hand.
//	Myojin of Seeing Winds has indestructible as long as it has a divinity counter on it.
//	Remove a divinity counter from Myojin of Seeing Winds: Draw a card for each permanent you control.
var MyojinOfSeeingWinds = newMyojinOfSeeingWinds

func newMyojinOfSeeingWinds() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Myojin of Seeing Winds",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Myojin of Seeing Winds has indestructible as long as it has a divinity counter on it.",
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
					Text: "Remove a divinity counter from Myojin of Seeing Winds: Draw a card for each permanent you control.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a divinity counter from Myojin of Seeing Winds",
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
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{Controller: game.ControllerYou}),
											}),
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
				game.EntersWithCountersIfReplacement("Myojin of Seeing Winds enters with a divinity counter on it if you cast it from your hand.", &game.Condition{
					EventPermanentWasCastFromControllerHand: true,
				}, game.CounterPlacement{Kind: counter.Divinity, Amount: 1}),
			},
			OracleText: `
			Myojin of Seeing Winds enters with a divinity counter on it if you cast it from your hand.
			Myojin of Seeing Winds has indestructible as long as it has a divinity counter on it.
			Remove a divinity counter from Myojin of Seeing Winds: Draw a card for each permanent you control.
		`,
		},
	}
}
