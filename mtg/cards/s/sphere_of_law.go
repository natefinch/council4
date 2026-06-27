package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphereOfLaw is the card definition for Sphere of Law.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	If a red source would deal damage to you, prevent 2 of that damage.
var SphereOfLaw = newSphereOfLaw()

func newSphereOfLaw() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sphere of Law",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If a red source would deal damage to you, prevent 2 of that damage.", &game.DamagePreventionSpec{Amount: 2, SourceColors: []color.Color{color.Red}, SourceTypes: nil, SourceControllerOpponent: false}),
			},
			OracleText: `
			If a red source would deal damage to you, prevent 2 of that damage.
		`,
		},
	}
}
