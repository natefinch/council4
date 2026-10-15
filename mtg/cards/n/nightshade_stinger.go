package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NightshadeStinger is the card definition for Nightshade Stinger.
//
// Type: Creature — Faerie Rogue
// Cost: {B}
//
// Oracle text:
//
//	Flying
//	This creature can't block.
var NightshadeStinger = newNightshadeStinger

func newNightshadeStinger() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Nightshade Stinger",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.CantBlockStaticBody,
			},
			OracleText: `
			Flying
			This creature can't block.
		`,
		},
	}
}
