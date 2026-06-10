package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Dragon Illusion
//
// Type: Token Creature — Dragon Illusion
//
// Oracle text:
//   Flying, haste

// DragonIllusionToken33b8ab7a1c2746f0a6827c9d7e9f43a6 is the card definition for Dragon Illusion.
var DragonIllusionToken33b8ab7a1c2746f0a6827c9d7e9f43a6 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Red),
	CardFace: game.CardFace{
		Name:      "Dragon Illusion",
		Colors:    []color.Color{color.Red},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Dragon, types.Illusion},
		Power:     opt.Val(game.PT{IsStar: true}),
		Toughness: opt.Val(game.PT{IsStar: true}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Flying, haste
		`,
	},
}
