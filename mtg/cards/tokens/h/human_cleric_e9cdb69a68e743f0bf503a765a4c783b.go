package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Human Cleric
//
// Type: Token Creature — Human Cleric
//
// Oracle text:
//   Lifelink, haste

// HumanClericTokene9cdb69a68e743f0bf503a765a4c783b is the card definition for Human Cleric.
var HumanClericTokene9cdb69a68e743f0bf503a765a4c783b = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.White, color.Red),
	CardFace: game.CardFace{
		Name:      "Human Cleric",
		Colors:    []color.Color{color.Red, color.White},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Human, types.Cleric},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.LifelinkStaticBody,
			game.HasteStaticBody,
		},
		OracleText: `
			Lifelink, haste
		`,
	},
}
