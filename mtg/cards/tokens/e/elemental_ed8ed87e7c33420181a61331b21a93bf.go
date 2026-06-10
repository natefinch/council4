package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Elemental
//
// Type: Token Creature — Elemental
//
// Oracle text:
//   Flying

// ElementalTokened8ed87e7c33420181a61331b21a93bf is the card definition for Elemental.
var ElementalTokened8ed87e7c33420181a61331b21a93bf = &game.CardDef{
	ColorIdentity: color.NewIdentity(color.Blue, color.Red),
	CardFace: game.CardFace{
		Name:      "Elemental",
		Colors:    []color.Color{color.Red, color.Blue},
		Types:     []types.Card{types.Creature},
		Subtypes:  []types.Sub{types.Elemental},
		Power:     opt.Val(game.PT{Value: 3}),
		Toughness: opt.Val(game.PT{Value: 3}),
		StaticAbilities: []game.StaticAbility{
			game.FlyingStaticBody,
		},
		OracleText: `
			Flying
		`,
	},
}
