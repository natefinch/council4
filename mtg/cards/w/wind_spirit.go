package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WindSpirit is the card definition for Wind Spirit.
//
// Type: Creature — Elemental Spirit
// Cost: {4}{U}
//
// Oracle text:
//
//	Flying
//	Menace (This creature can't be blocked except by two or more creatures.)
var WindSpirit = newWindSpirit

func newWindSpirit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wind Spirit",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Spirit},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.MenaceStaticBody,
			},
			OracleText: `
			Flying
			Menace (This creature can't be blocked except by two or more creatures.)
		`,
		},
	}
}
