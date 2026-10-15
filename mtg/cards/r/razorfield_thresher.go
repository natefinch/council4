package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RazorfieldThresher is the card definition for Razorfield Thresher.
//
// Type: Artifact Creature — Construct
// Cost: {7}
//
// Oracle text:
var RazorfieldThresher = newRazorfieldThresher

func newRazorfieldThresher() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Razorfield Thresher",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 4}),
		},
	}
}
