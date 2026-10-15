package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MinotaurAggressor is the card definition for Minotaur Aggressor.
//
// Type: Creature — Minotaur Berserker
// Cost: {6}{R}
//
// Oracle text:
//
//	First strike, haste
var MinotaurAggressor = newMinotaurAggressor

func newMinotaurAggressor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Minotaur Aggressor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Minotaur, types.Berserker},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			First strike, haste
		`,
		},
	}
}
