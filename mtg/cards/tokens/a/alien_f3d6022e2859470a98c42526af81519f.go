package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Alien is the card definition for Alien.
//
// Type: Token Creature — Alien
//
// Oracle text:
//   Haste
//   This token attacks each combat if able.

// AlienTokenf3d6022e2859470a98c42526af81519f is the card definition for Alien.
var AlienTokenf3d6022e2859470a98c42526af81519f = newAlienTokenf3d6022e2859470a98c42526af81519f()

func newAlienTokenf3d6022e2859470a98c42526af81519f() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:      "Alien",
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Alien},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Haste
			This token attacks each combat if able.
		`,
		},
	}
}
