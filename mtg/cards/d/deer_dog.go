package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DeerDog is the card definition for Deer-Dog.
//
// Type: Creature — Elk Dog
// Cost: {1}{R}
//
// Oracle text:
//
//	First strike (This creature deals combat damage before creatures without first strike.)
var DeerDog = newDeerDog

func newDeerDog() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Deer-Dog",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elk, types.Dog},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			OracleText: `
			First strike (This creature deals combat damage before creatures without first strike.)
		`,
		},
	}
}
