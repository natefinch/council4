package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Weird // Goblin
//
// Type: Token Creature — Weird // Token Creature — Goblin
// Face: Goblin — Token Creature — Goblin
//
// Oracle text:
//   Defender, flying

// WeirdToken9c56763e460342909e9f25f5eea8f4a5 is the card definition for Weird.
var WeirdToken9c56763e460342909e9f25f5eea8f4a5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Red),
	CardFace: game.CardFace{
		Name:      "Weird",
		Colors:    []color.Color{color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Weird},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.DefenderStaticBody,
			game.FlyingStaticBody,
		},
		OracleText: `
			Defender, flying
		`,
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Goblin",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Goblin},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	}),
}
