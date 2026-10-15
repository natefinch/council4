package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// StoneworkPuma is the card definition for Stonework Puma.
//
// Type: Artifact Creature — Cat Ally
// Cost: {3}
//
// Oracle text:
var StoneworkPuma = newStoneworkPuma

func newStoneworkPuma() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Stonework Puma",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Ally},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
