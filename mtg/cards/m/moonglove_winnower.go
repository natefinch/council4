package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoongloveWinnower is the card definition for Moonglove Winnower.
//
// Type: Creature — Elf Rogue
// Cost: {3}{B}
//
// Oracle text:
//
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var MoongloveWinnower = newMoongloveWinnower

func newMoongloveWinnower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Moonglove Winnower",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
