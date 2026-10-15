package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PhyrexianHulk is the card definition for Phyrexian Hulk.
//
// Type: Artifact Creature — Phyrexian Golem
// Cost: {6}
//
// Oracle text:
var PhyrexianHulk = newPhyrexianHulk

func newPhyrexianHulk() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Phyrexian Hulk",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Golem},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
