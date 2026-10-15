package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoldForgedSentinel is the card definition for Gold-Forged Sentinel.
//
// Type: Artifact Creature — Chimera
// Cost: {6}
//
// Oracle text:
//
//	Flying
var GoldForgedSentinel = newGoldForgedSentinel

func newGoldForgedSentinel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Gold-Forged Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Chimera},
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
