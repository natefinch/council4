package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FieldCreeper is the card definition for Field Creeper.
//
// Type: Artifact Creature — Scarecrow
// Cost: {2}
//
// Oracle text:
var FieldCreeper = newFieldCreeper

func newFieldCreeper() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Field Creeper",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Scarecrow},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
