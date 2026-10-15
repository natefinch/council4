package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// EbonyRhino is the card definition for Ebony Rhino.
//
// Type: Artifact Creature — Rhino
// Cost: {7}
//
// Oracle text:
//
//	Trample
var EbonyRhino = newEbonyRhino

func newEbonyRhino() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Ebony Rhino",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Rhino},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
			},
			OracleText: `
			Trample
		`,
		},
	}
}
