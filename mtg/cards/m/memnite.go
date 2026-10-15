package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Memnite is the card definition for Memnite.
//
// Type: Artifact Creature — Construct
// Cost: {0}
//
// Oracle text:
var Memnite = newMemnite

func newMemnite() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Memnite",
			ManaCost: opt.Val(cost.Mana{
				cost.O(0),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
