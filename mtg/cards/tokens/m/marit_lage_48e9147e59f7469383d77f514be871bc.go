package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Marit Lage
//
// Type: Token Legendary Creature — Avatar
//
// Oracle text:
//   Flying, indestructible

// MaritLageToken48e9147e59f7469383d77f514be871bc is the card definition for Marit Lage.
var MaritLageToken48e9147e59f7469383d77f514be871bc = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:       "Marit Lage",
		Colors:     []color.Color{color.Black},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Avatar},
		Power:      opt.Val(game.PT{Value: 20}),
		Toughness:  opt.Val(game.PT{Value: 20}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.IndestructibleStaticBody,
		},
		OracleText: `
			Flying, indestructible
		`,
	},
}
