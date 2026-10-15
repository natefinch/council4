package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RazortrapGorge is the card definition for Razortrap Gorge.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless a player has 13 or less life.
//	{T}: Add {B} or {R}.
var RazortrapGorge = newRazortrapGorge

func newRazortrapGorge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Razortrap Gorge",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless a player has 13 or less life.", &game.Condition{
					Negate:              true,
					AnyPlayerLifeAtMost: 13,
				}),
			},
			OracleText: `
			This land enters tapped unless a player has 13 or less life.
			{T}: Add {B} or {R}.
		`,
		},
	}
}
