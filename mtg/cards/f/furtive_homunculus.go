package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FurtiveHomunculus is the card definition for Furtive Homunculus.
//
// Type: Creature — Homunculus
// Cost: {1}{U}
//
// Oracle text:
//
//	Skulk (This creature can't be blocked by creatures with greater power.)
var FurtiveHomunculus = newFurtiveHomunculus

func newFurtiveHomunculus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Furtive Homunculus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Homunculus},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.SkulkStaticBody,
			},
			OracleText: `
			Skulk (This creature can't be blocked by creatures with greater power.)
		`,
		},
	}
}
