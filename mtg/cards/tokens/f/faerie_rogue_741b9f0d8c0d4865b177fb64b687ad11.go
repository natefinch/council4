package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Faerie Rogue
//
// Type: Token Creature — Faerie Rogue
//
// Oracle text:
//   Flying

// FaerieRogueToken741b9f0d8c0d4865b177fb64b687ad11 is the card definition for Faerie Rogue.
var FaerieRogueToken741b9f0d8c0d4865b177fb64b687ad11 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Faerie Rogue",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Faerie, types.Rogue},
		Power:     opt.Val(game.PT{Value: 1}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
