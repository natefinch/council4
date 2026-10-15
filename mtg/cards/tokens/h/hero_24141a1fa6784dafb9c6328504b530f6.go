package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Hero is the card definition for Hero.
//
// Type: Token Creature — Hero
//
// Oracle text:
//   Vigilance

// HeroToken24141a1fa6784dafb9c6328504b530f6 is the card definition for Hero.
var HeroToken24141a1fa6784dafb9c6328504b530f6 = newHeroToken24141a1fa6784dafb9c6328504b530f6()

func newHeroToken24141a1fa6784dafb9c6328504b530f6() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:      "Hero",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Hero},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			OracleText: `
			Vigilance
		`,
		},
	}
}
