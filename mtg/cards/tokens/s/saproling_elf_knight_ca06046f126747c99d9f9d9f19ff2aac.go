package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Saproling // Elf Knight
//
// Type: Token Creature — Saproling // Token Creature — Elf Knight
// Face: Elf Knight — Token Creature — Elf Knight
//
// Oracle text:
//   Elf Knight
//   Vigilance

// SaprolingTokenca06046f126747c99d9f9d9f19ff2aac is the card definition for Saproling.
var SaprolingTokenca06046f126747c99d9f9d9f19ff2aac = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Saproling",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Saproling},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Elf Knight",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elf, types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	}),
}
