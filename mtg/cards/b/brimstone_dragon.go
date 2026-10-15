package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrimstoneDragon is the card definition for Brimstone Dragon.
//
// Type: Creature — Dragon
// Cost: {6}{R}{R}
//
// Oracle text:
//
//	Flying, haste
var BrimstoneDragon = newBrimstoneDragon

func newBrimstoneDragon() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Brimstone Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Flying, haste
		`,
		},
	}
}
