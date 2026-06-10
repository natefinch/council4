package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Vampire
//
// Type: Token Creature — Vampire
//
// Oracle text:
//   Flying, lifelink

// VampireToken4a47367ebb5c42d9bc7fc79fc03e9ce1 is the card definition for Vampire.
var VampireToken4a47367ebb5c42d9bc7fc79fc03e9ce1 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Vampire",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Vampire},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.LifelinkStaticBody,
		},
		OracleText: `
			Flying, lifelink
		`,
	},
}
