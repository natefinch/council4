package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RustedSentinel is the card definition for Rusted Sentinel.
//
// Type: Artifact Creature — Golem
// Cost: {4}
//
// Oracle text:
//
//	This creature enters tapped.
var RustedSentinel = newRustedSentinel

func newRustedSentinel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Rusted Sentinel",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			OracleText: `
			This creature enters tapped.
		`,
		},
	}
}
