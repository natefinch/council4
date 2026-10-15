package z

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ZodiacRooster is the card definition for Zodiac Rooster.
//
// Type: Creature — Bird
// Cost: {1}{G}
//
// Oracle text:
//
//	Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)
var ZodiacRooster = newZodiacRooster

func newZodiacRooster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Zodiac Rooster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Bird},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.PlainswalkStaticBody,
			},
			OracleText: `
			Plainswalk (This creature can't be blocked as long as defending player controls a Plains.)
		`,
		},
	}
}
