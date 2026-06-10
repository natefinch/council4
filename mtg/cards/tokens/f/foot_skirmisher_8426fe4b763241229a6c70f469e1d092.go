package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Foot Skirmisher
//
// Type: Creature — Ninja
//
// Oracle text:
//   Flying (This creature can't be blocked except by creatures with flying or reach.)
//   This creature can't block.

// FootSkirmisherToken8426fe4b763241229a6c70f469e1d092 is the card definition for Foot Skirmisher.
var FootSkirmisherToken8426fe4b763241229a6c70f469e1d092 = &game.CardDef{
	CardFace: game.CardFace{
		Name:      "Foot Skirmisher",
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Ninja},
		Power:     opt.Val(game.PT{Value: 2}),
		Toughness: opt.Val(game.PT{Value: 1}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
			game.CantBlockStaticBody,
		},
		OracleText: `
			Flying (This creature can't be blocked except by creatures with flying or reach.)
			This creature can't block.
		`,
	},
}
