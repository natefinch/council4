package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HighbornGhoul is the card definition for Highborn Ghoul.
//
// Type: Creature — Zombie
// Cost: {B}{B}
//
// Oracle text:
//
//	Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
var HighbornGhoul = newHighbornGhoul

func newHighbornGhoul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Highborn Ghoul",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.IntimidateStaticBody,
			},
			OracleText: `
			Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
		`,
		},
	}
}
