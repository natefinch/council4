package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BattleBrawler is the card definition for Battle Brawler.
//
// Type: Creature — Orc Warrior
// Cost: {1}{B}
//
// Oracle text:
//
//	As long as you control a red or white permanent, this creature gets +1/+0 and has first strike.
var BattleBrawler = newBattleBrawler

func newBattleBrawler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Battle Brawler",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control a red or white permanent, this creature gets +1/+0 and has first strike.",
					Condition: opt.Val(game.Condition{
						Text: "As long as you control a red or white permanent",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{ColorsAny: []color.Color{color.Red, color.White}},
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
							AddKeywords:    []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			As long as you control a red or white permanent, this creature gets +1/+0 and has first strike.
		`,
		},
	}
}
