package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Phyrexian Hydra
//
// Type: Token Creature — Phyrexian Hydra
//
// Oracle text:
//   Reach

// PhyrexianHydraToken37412dd577b44166b0a4644f2c879e1b is the card definition for Phyrexian Hydra.
var PhyrexianHydraToken37412dd577b44166b0a4644f2c879e1b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Phyrexian Hydra",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Hydra},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.ReachStaticBody,
		},
		OracleText: `
			Reach
		`,
	},
}
