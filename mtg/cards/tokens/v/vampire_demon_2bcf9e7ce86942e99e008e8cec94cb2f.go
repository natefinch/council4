package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vampire Demon
//
// Type: Token Creature — Vampire Demon
//
// Oracle text:
//   Flying

// VampireDemonToken2bcf9e7ce86942e99e008e8cec94cb2f is the card definition for Vampire Demon.
var VampireDemonToken2bcf9e7ce86942e99e008e8cec94cb2f = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Black),
	CardFace: game.CardFace{
		Name:      "Vampire Demon",
		Colors:    []color.Color{color.Black, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Vampire, types.Demon},
		Power:     opt.Val(game.PT{Value: 4}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
