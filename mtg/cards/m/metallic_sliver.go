package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MetallicSliver is the card definition for Metallic Sliver.
//
// Type: Artifact Creature — Sliver
// Cost: {1}
//
// Oracle text:
var MetallicSliver = newMetallicSliver

func newMetallicSliver() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Metallic Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
