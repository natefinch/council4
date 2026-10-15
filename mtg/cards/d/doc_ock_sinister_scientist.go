package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DocOckSinisterScientist is the card definition for Doc Ock, Sinister Scientist.
//
// Type: Legendary Creature — Human Scientist Villain
// Cost: {4}{U}
//
// Oracle text:
//
//	As long as there are eight or more cards in your graveyard, Doc Ock has base power and toughness 8/8.
//	As long as you control another Villain, Doc Ock has hexproof. (He can't be the target of spells or abilities your opponents control.)
var DocOckSinisterScientist = newDocOckSinisterScientist

func newDocOckSinisterScientist() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Doc Ock, Sinister Scientist",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:     []color.Color{color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Scientist, types.Villain},
			Power:      opt.Val(game.PT{Value: 4}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as there are eight or more cards in your graveyard, Doc Ock has base power and toughness 8/8.",
					Condition: opt.Val(game.Condition{
						Text: "As long as there are eight or more cards in your graveyard",
						Aggregates: []game.AggregateComparison{
							game.AggregateComparison{
								Aggregate: game.AggregateControllerGraveyardCardCount,
								Op:        compare.GreaterOrEqual,
								Value:     8,
							},
						},
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessSet,
							SetPower: opt.Val(game.PT{
								Value: 8,
							}),
							SetToughness: opt.Val(game.PT{
								Value: 8,
							}),
						},
					},
				},
				game.StaticAbility{
					Text: "As long as you control another Villain, Doc Ock has hexproof. (He can't be the target of spells or abilities your opponents control.)",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control another Villain",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Villain}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Hexproof},
						},
					},
				},
			},
			OracleText: `
			As long as there are eight or more cards in your graveyard, Doc Ock has base power and toughness 8/8.
			As long as you control another Villain, Doc Ock has hexproof. (He can't be the target of spells or abilities your opponents control.)
		`,
		},
	}
}
