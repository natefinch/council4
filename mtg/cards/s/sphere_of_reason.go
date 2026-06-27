package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphereOfReason is the card definition for Sphere of Reason.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	If a blue source would deal damage to you, prevent 2 of that damage.
var SphereOfReason = newSphereOfReason()

func newSphereOfReason() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sphere of Reason",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a blue source would deal damage to you, prevent 2 of that damage.", &game.DamagePreventionSpec{Amount: 2, SourceColors: []color.Color{color.Blue}, SourceTypes: nil, SourceControllerOpponent: false}),
			},
			OracleText: `
			If a blue source would deal damage to you, prevent 2 of that damage.
		`,
		},
	}
}
