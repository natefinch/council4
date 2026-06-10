package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Foot Disciple
//
// Type: Creature — Ninja
//
// Oracle text:
//   This creature can't block.

// FootDiscipleTokencbf0b7c6c79c43bcba8ab3c6b12a4433 is the card definition for Foot Disciple.
var FootDiscipleTokencbf0b7c6c79c43bcba8ab3c6b12a4433 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Foot Disciple",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ninja},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.CantBlockStaticBody,
		},
		OracleText: `
			This creature can't block.
		`,
	},
}
