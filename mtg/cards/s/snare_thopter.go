package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SnareThopter is the card definition for Snare Thopter.
//
// Type: Artifact Creature — Thopter
// Cost: {4}
//
// Oracle text:
//
//	Flying, haste
var SnareThopter = newSnareThopter

func newSnareThopter() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Snare Thopter",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Thopter},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			OracleText: `
			Flying, haste
		`,
		},
	}
}
