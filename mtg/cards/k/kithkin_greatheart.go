package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KithkinGreatheart is the card definition for Kithkin Greatheart.
//
// Type: Creature — Kithkin Soldier
// Cost: {1}{W}
//
// Oracle text:
//
//	As long as you control a Giant, this creature gets +1/+1 and has first strike.
var KithkinGreatheart = newKithkinGreatheart

func newKithkinGreatheart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Kithkin Greatheart",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Kithkin, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a Giant, this creature gets +1/+1 and has first strike.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Giant",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Giant}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     1,
							ToughnessDelta: 1,
						},
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			As long as you control a Giant, this creature gets +1/+1 and has first strike.
		`,
		},
	}
}
