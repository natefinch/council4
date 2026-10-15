package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WickerWitch is the card definition for Wicker Witch.
//
// Type: Artifact Creature — Scarecrow
// Cost: {3}
//
// Oracle text:
var WickerWitch = newWickerWitch

func newWickerWitch() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Wicker Witch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
