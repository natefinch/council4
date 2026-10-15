package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EldraziAggressor is the card definition for Eldrazi Aggressor.
//
// Type: Creature — Eldrazi Drone
// Cost: {2}{R}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	This creature has haste as long as you control another colorless creature.
var EldraziAggressor = newEldraziAggressor

func newEldraziAggressor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Eldrazi Aggressor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Drone},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.StaticAbility{
					Text: "This creature has haste as long as you control another colorless creature.",
					Condition: opt.Val(game.Condition{
						Text: "as long as you control another colorless creature",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Colorless: true, ExcludeSource: true},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerAbility,
							AddKeywords:    []game.Keyword{game.Haste},
						},
					},
				},
			},
			OracleText: `
			Devoid (This card has no color.)
			This creature has haste as long as you control another colorless creature.
		`,
		},
	}
}
