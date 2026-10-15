package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DemonSJester is the card definition for Demon's Jester.
//
// Type: Creature — Imp
// Cost: {3}{B}
//
// Oracle text:
//
//	Flying
//	Hellbent — This creature gets +2/+1 as long as you have no cards in hand.
var DemonSJester = newDemonSJester

func newDemonSJester() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Demon's Jester",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Imp},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "Hellbent — This creature gets +2/+1 as long as you have no cards in hand.",
					Condition: opt.Val(game.Condition{
						Text:                "as long as you have no cards in hand",
						ControllerHandEmpty: true,
						SourceCounterKind:   counter.PlusOnePlusOne,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:       game.DurationPermanent,
							AffectedSource: true,
							Layer:          game.LayerPowerToughnessModify,
							PowerDelta:     2,
							ToughnessDelta: 1,
						},
					},
				},
			},
			OracleText: `
			Flying
			Hellbent — This creature gets +2/+1 as long as you have no cards in hand.
		`,
		},
	}
}
