package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Arachnoid is the card definition for Arachnoid.
//
// Type: Artifact Creature — Spider
// Cost: {6}
//
// Oracle text:
//
//	Reach (This creature can block creatures with flying.)
var Arachnoid = newArachnoid

func newArachnoid() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Arachnoid",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Spider},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.ReachStaticBody,
			},
			OracleText: `
			Reach (This creature can block creatures with flying.)
		`,
		},
	}
}
