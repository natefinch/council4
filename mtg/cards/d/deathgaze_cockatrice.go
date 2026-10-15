package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeathgazeCockatrice is the card definition for Deathgaze Cockatrice.
//
// Type: Creature — Cockatrice
// Cost: {2}{B}{B}
//
// Oracle text:
//
//	Flying
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var DeathgazeCockatrice = newDeathgazeCockatrice

func newDeathgazeCockatrice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Deathgaze Cockatrice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cockatrice},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Flying
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
