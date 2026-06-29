package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OrbsOfWarding is the card definition for Orbs of Warding.
//
// Type: Artifact
// Cost: {5}
//
// Oracle text:
//
//	You have hexproof. (You can't be the target of spells or abilities your opponents control.)
//	If a creature would deal damage to you, prevent 1 of that damage.
var OrbsOfWarding = newOrbsOfWarding()

func newOrbsOfWarding() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Orbs of Warding",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.PlayerHexproofStaticBody,
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a creature would deal damage to you, prevent 1 of that damage.", &game.DamagePreventionSpec{Amount: 1, SourceColors: nil, SourceTypes: []types.Card{types.Creature}, SourceControllerOpponent: false}),
			},
			OracleText: `
			You have hexproof. (You can't be the target of spells or abilities your opponents control.)
			If a creature would deal damage to you, prevent 1 of that damage.
		`,
		},
	}
}
