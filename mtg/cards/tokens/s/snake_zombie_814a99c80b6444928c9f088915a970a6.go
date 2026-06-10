package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Snake // Zombie
//
// Type: Token Creature — Snake // Token Creature — Zombie
// Face: Zombie — Token Creature — Zombie
//
// Oracle text:
//   Deathtouch

// SnakeToken814a99c80b6444928c9f088915a970a6 is the card definition for Snake.
var SnakeToken814a99c80b6444928c9f088915a970a6 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Snake",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
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
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Zombie",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Zombie},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
	}),
}
