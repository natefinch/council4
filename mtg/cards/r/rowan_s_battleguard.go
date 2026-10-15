package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RowanSBattleguard is the card definition for Rowan's Battleguard.
//
// Type: Creature — Human Knight
// Cost: {3}{R}
//
// Oracle text:
//
//	First strike
//	As long as you control a Rowan planeswalker, this creature gets +3/+0.
var RowanSBattleguard = newRowanSBattleguard

func newRowanSBattleguard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Rowan's Battleguard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Knight},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.StaticAbility{
					Text: "As long as you control a Rowan planeswalker, this creature gets +3/+0.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a Rowan planeswalker",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Rowan}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     3,
						},
					},
				},
			},
			OracleText: `
			First strike
			As long as you control a Rowan planeswalker, this creature gets +3/+0.
		`,
		},
	}
}
