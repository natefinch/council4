package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GlassGolem is the card definition for Glass Golem.
//
// Type: Artifact Creature — Golem
// Cost: {5}
//
// Oracle text:
var GlassGolem = newGlassGolem

func newGlassGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Glass Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
