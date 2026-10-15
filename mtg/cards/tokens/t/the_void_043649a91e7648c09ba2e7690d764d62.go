package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheVoid is the card definition for The Void.
//
// Type: Token Legendary Creature — Horror Villain
//
// Oracle text:
//   Flying, indestructible
//   The Void attacks each combat if able.

// TheVoidToken043649a91e7648c09ba2e7690d764d62 is the card definition for The Void.
var TheVoidToken043649a91e7648c09ba2e7690d764d62 = newTheVoidToken043649a91e7648c09ba2e7690d764d62()

func newTheVoidToken043649a91e7648c09ba2e7690d764d62() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "The Void",
			Colors:     []color.Color{color.Black},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Horror, types.Villain},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.IndestructibleStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Flying, indestructible
			The Void attacks each combat if able.
		`,
		},
	}
}
