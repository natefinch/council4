package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScoriaCat is the card definition for Scoria Cat.
//
// Type: Creature — Cat
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	This creature gets +3/+3 as long as you control no untapped lands.
var ScoriaCat = newScoriaCat

func newScoriaCat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Scoria Cat",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature gets +3/+3 as long as you control no untapped lands.",
					Condition: opt.Val(game.Condition{
						Text:   "as long as you control no untapped lands",
						Negate: true,
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Tapped: game.TriFalse},
							MinCount:  1,
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
							ToughnessDelta: 3,
						},
					},
				},
			},
			OracleText: `
			This creature gets +3/+3 as long as you control no untapped lands.
		`,
		},
	}
}
