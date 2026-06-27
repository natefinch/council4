package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SphereOfPurity is the card definition for Sphere of Purity.
//
// Type: Enchantment
// Cost: {3}{W}
//
// Oracle text:
//
//	If an artifact would deal damage to you, prevent 1 of that damage.
var SphereOfPurity = newSphereOfPurity()

func newSphereOfPurity() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Sphere of Purity",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors: []color.Color{color.White},
			Types:  []types.Card{types.Enchantment},
			ReplacementAbilities: []game.ReplacementAbility{
				game.DamagePreventionReplacement("If an artifact would deal damage to you, prevent 1 of that damage.", &game.DamagePreventionSpec{Amount: 1, SourceColors: nil, SourceTypes: []types.Card{types.Artifact}, SourceControllerOpponent: false}),
			},
			OracleText: `
			If an artifact would deal damage to you, prevent 1 of that damage.
		`,
		},
	}
}
