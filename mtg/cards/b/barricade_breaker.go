package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BarricadeBreaker is the card definition for Barricade Breaker.
//
// Type: Artifact Creature — Juggernaut
// Cost: {7}
//
// Oracle text:
//
//	Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
//	This creature attacks each combat if able.
var BarricadeBreaker = newBarricadeBreaker

func newBarricadeBreaker() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Barricade Breaker",
			ManaCost: opt.Val(cost.Mana{
				cost.O(7),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Juggernaut},
			Power:     opt.Val(game.PT{Value: 7}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.ImproviseStaticBody,
				game.MustAttackStaticBody,
			},
			OracleText: `
			Improvise (Your artifacts can help cast this spell. Each artifact you tap after you're done activating mana abilities pays for {1}.)
			This creature attacks each combat if able.
		`,
		},
	}
}
