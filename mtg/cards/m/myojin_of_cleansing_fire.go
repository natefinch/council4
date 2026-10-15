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

// MyojinOfCleansingFire is the card definition for Myojin of Cleansing Fire.
//
// Type: Legendary Creature — Spirit
// Cost: {5}{W}{W}{W}
//
// Oracle text:
//
//	Myojin of Cleansing Fire enters with a divinity counter on it if you cast it from your hand.
//	Myojin of Cleansing Fire has indestructible as long as it has a divinity counter on it.
//	Remove a divinity counter from Myojin of Cleansing Fire: Destroy all other creatures.
var MyojinOfCleansingFire = newMyojinOfCleansingFire

func newMyojinOfCleansingFire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Myojin of Cleansing Fire",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.W,
				cost.W,
				cost.W,
			}),
			Colors:     []color.Color{color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Myojin of Cleansing Fire has indestructible as long as it has a divinity counter on it.",
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
					Text: "Remove a divinity counter from Myojin of Cleansing Fire: Destroy all other creatures.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a divinity counter from Myojin of Cleansing Fire",
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
										Primitive: game.Destroy{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludeSource: true}),
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
				game.EntersWithCountersIfReplacement("Myojin of Cleansing Fire enters with a divinity counter on it if you cast it from your hand.", &game.Condition{
					EventPermanentWasCastFromControllerHand: true,
				}, game.CounterPlacement{Kind: counter.Divinity, Amount: 1}),
			},
			OracleText: `
			Myojin of Cleansing Fire enters with a divinity counter on it if you cast it from your hand.
			Myojin of Cleansing Fire has indestructible as long as it has a divinity counter on it.
			Remove a divinity counter from Myojin of Cleansing Fire: Destroy all other creatures.
		`,
		},
	}
}
