package y

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// YellowScarvesTroops is the card definition for Yellow Scarves Troops.
//
// Type: Creature — Human Soldier
// Cost: {1}{R}
//
// Oracle text:
//
//	This creature can't block.
var YellowScarvesTroops = newYellowScarvesTroops

func newYellowScarvesTroops() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Yellow Scarves Troops",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.CantBlockStaticBody,
			},
			OracleText: `
			This creature can't block.
		`,
		},
	}
}
