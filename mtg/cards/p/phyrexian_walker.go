package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianWalker is the card definition for Phyrexian Walker.
//
// Type: Artifact Creature — Phyrexian Construct
// Cost: {0}
//
// Oracle text:
var PhyrexianWalker = newPhyrexianWalker

func newPhyrexianWalker() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Phyrexian Walker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Construct},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 3}),
		},
	}
}
