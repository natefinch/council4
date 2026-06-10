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
//   Lifelink

// PhyrexianHydraToken4fef22b64ac94df2969a09abf1cd2cbb is the card definition for Phyrexian Hydra.
var PhyrexianHydraToken4fef22b64ac94df2969a09abf1cd2cbb = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:      "Phyrexian Hydra",
		Colors:    []color.Color{color.Green, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Phyrexian, types.Hydra},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
		},
		OracleText: `
			Lifelink
		`,
	},
}
