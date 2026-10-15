package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ChardalynDragon is the card definition for Chardalyn Dragon.
//
// Type: Artifact Creature — Dragon
// Cost: {6}
//
// Oracle text:
//
//	Flying
var ChardalynDragon = newChardalynDragon

func newChardalynDragon() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Chardalyn Dragon",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Dragon},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			OracleText: `
			Flying
		`,
		},
	}
}
