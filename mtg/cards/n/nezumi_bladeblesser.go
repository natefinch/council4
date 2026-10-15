package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NezumiBladeblesser is the card definition for Nezumi Bladeblesser.
//
// Type: Creature — Rat Samurai
// Cost: {2}{B}
//
// Oracle text:
//
//	This creature has deathtouch as long as you control an artifact.
//	This creature has menace as long as you control an enchantment. (It can't be blocked except by two or more creatures.)
var NezumiBladeblesser = newNezumiBladeblesser

func newNezumiBladeblesser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nezumi Bladeblesser",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rat, types.Samurai},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "This creature has deathtouch as long as you control an artifact.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control an artifact",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Deathtouch},
						},
					},
				},
				game.StaticAbility{
					Text: "This creature has menace as long as you control an enchantment. (It can't be blocked except by two or more creatures.)",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control an enchantment",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Menace},
						},
					},
				},
			},
			OracleText: `
			This creature has deathtouch as long as you control an artifact.
			This creature has menace as long as you control an enchantment. (It can't be blocked except by two or more creatures.)
		`,
		},
	}
}
