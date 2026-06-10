package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Knight
//
// Type: Token Creature — Knight
//
// Oracle text:
//   Protection from red

// KnightTokenc20063be7f7a4c5d9a21431906e62e68 is the card definition for Knight.
var KnightTokenc20063be7f7a4c5d9a21431906e62e68 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White),
	CardFace: game.CardFace{
		Name:      "Knight",
		Colors:    []color.Color{color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Knight},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.ProtectionFromColorsStaticAbility(color.Red),
		},
		OracleText: `
			Protection from red
		`,
	},
}
