package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CustodianOfTheTrove is the card definition for Custodian of the Trove.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	Defender
//	This creature enters tapped.
var CustodianOfTheTrove = newCustodianOfTheTrove

func newCustodianOfTheTrove() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Custodian of the Trove",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.DefenderStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This creature enters tapped."),
			},
			OracleText: `
			Defender
			This creature enters tapped.
		`,
		},
	}
}
