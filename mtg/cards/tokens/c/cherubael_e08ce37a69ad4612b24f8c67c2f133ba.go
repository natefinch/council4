package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cherubael
//
// Type: Token Legendary Creature — Demon
//
// Oracle text:
//   Flying

// CherubaelTokene08ce37a69ad4612b24f8c67c2f133ba is the card definition for Cherubael.
var CherubaelTokene08ce37a69ad4612b24f8c67c2f133ba = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:       "Cherubael",
		Colors:     []color.Color{color.Black},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Demon},
		Power:      opt.Val(game.PT{Value: 4}),
		Toughness:  opt.Val(game.PT{Value: 4}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
