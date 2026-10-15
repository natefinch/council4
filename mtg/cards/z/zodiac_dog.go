package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZodiacDog is the card definition for Zodiac Dog.
//
// Type: Creature — Dog
// Cost: {2}{R}
//
// Oracle text:
//
//	Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
var ZodiacDog = newZodiacDog

func newZodiacDog() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Zodiac Dog",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dog},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.MountainwalkStaticBody,
			},
			OracleText: `
			Mountainwalk (This creature can't be blocked as long as defending player controls a Mountain.)
		`,
		},
	}
}
