package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ProtectionOfTheHekma is the card definition for Protection of the Hekma.
//
// Type: Enchantment
// Cost: {4}{W}
//
// Oracle text:
//
//	If a source an opponent controls would deal damage to you, prevent 1 of that damage.
var ProtectionOfTheHekma = newProtectionOfTheHekma

func newProtectionOfTheHekma() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Protection of the Hekma",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a source an opponent controls would deal damage to you, prevent 1 of that damage.", &game.DamagePreventionSpec{Amount: 1, SourceColors: nil, SourceTypes: nil, SourceControllerOpponent: true}),
			},
			OracleText: `
			If a source an opponent controls would deal damage to you, prevent 1 of that damage.
		`,
		},
	}
}
