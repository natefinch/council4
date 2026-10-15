package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PardicWanderer is the card definition for Pardic Wanderer.
//
// Type: Artifact Creature — Golem
// Cost: {6}
//
// Oracle text:
//
//	Trample
var PardicWanderer = newPardicWanderer

func newPardicWanderer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Pardic Wanderer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 5}),
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
