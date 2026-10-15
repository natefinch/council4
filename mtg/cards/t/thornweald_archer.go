package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ThornwealdArcher is the card definition for Thornweald Archer.
//
// Type: Creature — Elf Archer
// Cost: {1}{G}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var ThornwealdArcher = newThornwealdArcher

func newThornwealdArcher() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Thornweald Archer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Archer},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
				game.DeathtouchStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
			Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
		`,
		},
	}
}
