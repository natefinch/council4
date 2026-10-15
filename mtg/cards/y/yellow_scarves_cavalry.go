package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YellowScarvesCavalry is the card definition for Yellow Scarves Cavalry.
//
// Type: Creature — Human Soldier
// Cost: {1}{R}
//
// Oracle text:
//
//	Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
//	This creature can't block.
var YellowScarvesCavalry = newYellowScarvesCavalry

func newYellowScarvesCavalry() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Yellow Scarves Cavalry",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.HorsemanshipStaticBody,
				game.CantBlockStaticBody,
			},
			OracleText: `
			Horsemanship (This creature can't be blocked except by creatures with horsemanship.)
			This creature can't block.
		`,
		},
	}
}
