package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NayaHushblade is the card definition for Naya Hushblade.
//
// Type: Creature — Elf Rogue
// Cost: {R/W}{G}
//
// Oracle text:
//
//	As long as you control another multicolored permanent, this creature gets +1/+1 and has shroud. (It can't be the target of spells or abilities.)
var NayaHushblade = newNayaHushblade

func newNayaHushblade() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Naya Hushblade",
			ManaCost: opt.Val(cost.Mana{
				cost.HybridMana(mana.R, mana.W),
				cost.G,
			}),
			Colors:    []color.Color{color.Green, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as you control another multicolored permanent, this creature gets +1/+1 and has shroud. (It can't be the target of spells or abilities.)",
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
							AddKeywords:    []game.Keyword{game.Shroud},
						},
					},
				},
			},
			OracleText: `
			As long as you control another multicolored permanent, this creature gets +1/+1 and has shroud. (It can't be the target of spells or abilities.)
		`,
		},
	}
}
