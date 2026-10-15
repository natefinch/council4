package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JundHackblade is the card definition for Jund Hackblade.
//
// Type: Creature — Goblin Berserker
// Cost: {B/G}{R}
//
// Oracle text:
//
//	As long as you control another multicolored permanent, this creature gets +1/+1 and has haste.
var JundHackblade = newJundHackblade

func newJundHackblade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Jund Hackblade",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.B, mana.G),
				cost.R,
			}),
			Colors:    []color.Color{color.Black, color.Green, color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Berserker},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control another multicolored permanent, this creature gets +1/+1 and has haste.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control another multicolored permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{Multicolored: true, ExcludeSource: true},
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
							AddKeywords:    []game.Keyword{game.Haste},
						},
					},
				},
			},
			OracleText: `
			As long as you control another multicolored permanent, this creature gets +1/+1 and has haste.
		`,
		},
	}
}
