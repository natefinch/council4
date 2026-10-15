package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Flutterfox is the card definition for Flutterfox.
//
// Type: Creature — Fox
// Cost: {1}{W}
//
// Oracle text:
//
//	As long as you control an artifact or enchantment, this creature has flying.
var Flutterfox = newFlutterfox

func newFlutterfox() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Flutterfox",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Fox},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control an artifact or enchantment, this creature has flying.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control an artifact or enchantment",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Flying},
						},
					},
				},
			},
			OracleText: `
			As long as you control an artifact or enchantment, this creature has flying.
		`,
		},
	}
}
