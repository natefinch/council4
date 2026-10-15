package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// IronBarbHellion is the card definition for Iron-Barb Hellion.
//
// Type: Creature — Hellion Beast
// Cost: {5}{R}
//
// Oracle text:
//
//	Haste
//	This creature can't block.
var IronBarbHellion = newIronBarbHellion

func newIronBarbHellion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Iron-Barb Hellion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hellion, types.Beast},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.CantBlockStaticBody,
			},
			OracleText: `
			Haste
			This creature can't block.
		`,
		},
	}
}
