package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HexplateGolem is the card definition for Hexplate Golem.
//
// Type: Artifact Creature — Golem
// Cost: {7}
//
// Oracle text:
var HexplateGolem = newHexplateGolem

func newHexplateGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hexplate Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
		},
	}
}
