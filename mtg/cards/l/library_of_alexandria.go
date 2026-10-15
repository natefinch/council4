package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LibraryOfAlexandria is the card definition for Library of Alexandria.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{T}: Draw a card. Activate only if you have exactly seven cards in hand.
var LibraryOfAlexandria = newLibraryOfAlexandria

func newLibraryOfAlexandria() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Library of Alexandria",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Draw a card. Activate only if you have exactly seven cards in hand.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Text: "only if you have exactly seven cards in hand",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerHandSize,
								Op:        compare.Equal,
								Value:     7,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
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
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{T}: Draw a card. Activate only if you have exactly seven cards in hand.
		`,
		},
	}
}
