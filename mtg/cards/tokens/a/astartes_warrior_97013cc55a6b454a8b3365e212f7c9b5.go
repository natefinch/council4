package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Astartes Warrior
//
// Type: Token Creature — Astartes Warrior
//
// Oracle text:
//   Menace

// AstartesWarriorToken97013cc55a6b454a8b3365e212f7c9b5 is the card definition for Astartes Warrior.
var AstartesWarriorToken97013cc55a6b454a8b3365e212f7c9b5 = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Black),
	CardFace: game.CardFace{
		Name:      "Astartes Warrior",
		Colors:    []color.Color{color.Black},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Astartes, types.Warrior},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 2}),
		StaticAbilities: []game.StaticAbility{
			game.MenaceStaticBody,
		},
		OracleText: `
			Menace
		`,
	},
}
