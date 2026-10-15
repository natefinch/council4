package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RancidRats is the card definition for Rancid Rats.
//
// Type: Creature — Zombie Rat
// Cost: {1}{B}
//
// Oracle text:
//
//	Skulk (This creature can't be blocked by creatures with greater power.)
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var RancidRats = newRancidRats

func newRancidRats() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Rancid Rats",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Zombie, types.Rat},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.SkulkStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Skulk (This creature can't be blocked by creatures with greater power.)
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
