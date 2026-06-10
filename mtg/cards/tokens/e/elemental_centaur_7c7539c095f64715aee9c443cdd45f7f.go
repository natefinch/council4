package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental // Centaur
//
// Type: Token Creature — Elemental // Token Creature — Centaur
// Face: Centaur — Token Creature — Centaur
//
// Oracle text:
//   Vigilance

// ElementalToken7c7539c095f64715aee9c443cdd45f7f is the card definition for Elemental.
var ElementalToken7c7539c095f64715aee9c443cdd45f7f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 8}),
		Toughness: opt.Val(game.PT{Value: 8}),
		StaticAbilities: []game.StaticAbility{
			game.VigilanceStaticBody,
		},
		OracleText: `
			Vigilance
		`,
	},
	Layout: game.LayoutDoubleFacedToken,
	Back: opt.Val(game.CardFace{
		Name:      "Centaur",
		Colors:    []color.Color{color.Green},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Centaur},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
	}),
}
