package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Graveborn
//
// Type: Token Creature — Graveborn
//
// Oracle text:
//   Haste

// GravebornToken3b2ccc7967d6437d9696fbd887c7d8b6 is the card definition for Graveborn.
var GravebornToken3b2ccc7967d6437d9696fbd887c7d8b6 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black, color.Red),
	CardFace: game.CardFace{
		Name:      "Graveborn",
		Colors:    []color.Color{color.Black, color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Graveborn},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.HasteStaticBody,
		},
		OracleText: `
			Haste
		`,
	},
}
