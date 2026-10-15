package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Villain is the card definition for Villain.
//
// Type: Token Creature — Villain
//
// Oracle text:
//   Menace

// VillainToken4ca0433808e4411ea89136ab9c320236 is the card definition for Villain.
var VillainToken4ca0433808e4411ea89136ab9c320236 = newVillainToken4ca0433808e4411ea89136ab9c320236()

func newVillainToken4ca0433808e4411ea89136ab9c320236() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:      "Villain",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
			OracleText: `
			Menace
		`,
		},
	}
}
