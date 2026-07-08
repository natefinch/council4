package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JaceSSentinel is the card definition for Jace's Sentinel.
//
// Type: Creature — Merfolk Warrior
// Cost: {1}{U}
//
// Oracle text:
//
//	As long as you control a Jace planeswalker, this creature gets +1/+0 and can't be blocked.
var JaceSSentinel = newJaceSSentinel

func newJaceSSentinel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Jace's Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Merfolk, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}, SubtypesAny: []types.Sub{types.Sub("Jace")}},
						}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							AffectedSource: true,
							PowerDelta:     1,
						},
					},
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantBeBlocked,
							AffectedSource: true,
						},
					},
				},
			},
			OracleText: `
			As long as you control a Jace planeswalker, this creature gets +1/+0 and can't be blocked.
		`,
		},
	}
}
