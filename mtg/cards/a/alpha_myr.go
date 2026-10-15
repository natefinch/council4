package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AlphaMyr is the card definition for Alpha Myr.
//
// Type: Artifact Creature — Myr
// Cost: {2}
//
// Oracle text:
var AlphaMyr = newAlphaMyr

func newAlphaMyr() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Alpha Myr",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Myr},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
