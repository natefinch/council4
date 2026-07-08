package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrzaSArmor is the card definition for Urza's Armor.
//
// Type: Artifact
// Cost: {6}
//
// Oracle text:
//
//	If a source would deal damage to you, prevent 1 of that damage.
var UrzaSArmor = newUrzaSArmor

func newUrzaSArmor() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Urza's Armor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Types: []types.Card{types.Artifact},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a source would deal damage to you, prevent 1 of that damage.", &game.DamagePreventionSpec{Amount: 1, SourceColors: nil, SourceTypes: nil, SourceControllerOpponent: false}),
			},
			OracleText: `
			If a source would deal damage to you, prevent 1 of that damage.
		`,
		},
	}
}
