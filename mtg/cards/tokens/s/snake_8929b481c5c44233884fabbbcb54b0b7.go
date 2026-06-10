package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Snake
//
// Type: Token Enchantment Creature — Snake
//
// Oracle text:
//   Deathtouch

// SnakeToken8929b481c5c44233884fabbbcb54b0b7 is the card definition for Snake.
var SnakeToken8929b481c5c44233884fabbbcb54b0b7 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Green),
	CardFace: game.CardFace{
		Name:      "Snake",
		Colors:    []color.Color{color.Black, color.Green},
		Types:     []types.Card{types.Enchantment, types.Creature},
		Subtypes:  []types.Sub{types.Snake},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
