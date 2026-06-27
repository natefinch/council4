package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LoyalPegasus is the card definition for Loyal Pegasus.
//
// Type: Creature — Pegasus
// Cost: {W}
//
// Oracle text:
//
//	Flying
//	This creature can't attack or block alone.
var LoyalPegasus = newLoyalPegasus()

func newLoyalPegasus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Loyal Pegasus",
			ManaCost: opt.Val(cost.Mana{
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Pegasus},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.CantAttackOrBlockAloneStaticBody,
			},
			OracleText: `
			Flying
			This creature can't attack or block alone.
		`,
		},
	}
}
