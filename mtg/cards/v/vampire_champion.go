package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// VampireChampion is the card definition for Vampire Champion.
//
// Type: Creature — Vampire Soldier
// Cost: {3}{B}
//
// Oracle text:
//
//	Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
var VampireChampion = newVampireChampion

func newVampireChampion() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Vampire Champion",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vampire, types.Soldier},
			Power:     opt.Val(game.PT{Value: 3}),
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
