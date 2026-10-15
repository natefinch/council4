package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnarlingWarg is the card definition for Snarling Warg.
//
// Type: Creature — Wolf
// Cost: {3}{B}
//
// Oracle text:
//
//	Menace (This creature can't be blocked except by two or more creatures.)
//	As long as you control a Goblin or Orc, this creature gets +1/+0.
var SnarlingWarg = newSnarlingWarg

func newSnarlingWarg() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Snarling Warg",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wolf},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
				game.StaticAbility{
					Text: "As long as you control a Goblin or Orc, this creature gets +1/+0.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Goblin or Orc",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Goblin, types.Orc}},
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
					},
				},
			},
			OracleText: `
			Menace (This creature can't be blocked except by two or more creatures.)
			As long as you control a Goblin or Orc, this creature gets +1/+0.
		`,
		},
	}
}
