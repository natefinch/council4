package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SliverConstruct is the card definition for Sliver Construct.
//
// Type: Artifact Creature — Sliver Construct
// Cost: {3}
//
// Oracle text:
var SliverConstruct = newSliverConstruct

func newSliverConstruct() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Sliver Construct",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Sliver, types.Construct},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
