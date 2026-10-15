package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HungryRidgewolf is the card definition for Hungry Ridgewolf.
//
// Type: Creature — Wolf
// Cost: {1}{R}
//
// Oracle text:
//
//	As long as you control another Wolf or Werewolf, this creature gets +1/+0 and has trample.
var HungryRidgewolf = newHungryRidgewolf

func newHungryRidgewolf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Hungry Ridgewolf",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control another Wolf or Werewolf, this creature gets +1/+0 and has trample.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control another Wolf or Werewolf",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Wolf, types.Werewolf}, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Trample},
						},
					},
				},
			},
			OracleText: `
			As long as you control another Wolf or Werewolf, this creature gets +1/+0 and has trample.
		`,
		},
	}
}
