package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BronzeSable is the card definition for Bronze Sable.
//
// Type: Artifact Creature — Sable
// Cost: {2}
//
// Oracle text:
var BronzeSable = newBronzeSable

func newBronzeSable() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Bronze Sable",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Sable},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
