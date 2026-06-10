package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Beast
//
// Type: Token Creature — Beast
//
// Oracle text:
//   Deathtouch

// BeastToken3c27a325449044e6aa9ab591697486c8 is the card definition for Beast.
var BeastToken3c27a325449044e6aa9ab591697486c8 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Beast",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Beast},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.DeathtouchStaticBody,
		},
		OracleText: `
			Deathtouch
		`,
	},
}
