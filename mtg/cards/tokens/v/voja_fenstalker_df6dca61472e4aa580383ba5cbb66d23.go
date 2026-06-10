package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Voja Fenstalker
//
// Type: Token Legendary Creature — Wolf
//
// Oracle text:
//   Trample

// VojaFenstalkerTokendf6dca61472e4aa580383ba5cbb66d23 is the card definition for Voja Fenstalker.
var VojaFenstalkerTokendf6dca61472e4aa580383ba5cbb66d23 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Green),
	CardFace: game.CardFace{
		Name:       "Voja Fenstalker",
		Colors:     []color.Color{color.Green, color.White},
		Supertypes: []types.Super{types.Legendary},
		Types:      []types.Card{types.Creature},
		Subtypes:   []types.Sub{types.Wolf},
		Power:      opt.Val(game.PT{Value: 5}),
		Toughness:  opt.Val(game.PT{Value: 5}),
		StaticAbilities: []game.StaticAbility{
			game.TrampleStaticBody,
		},
		OracleText: `
			Trample
		`,
	},
}
