package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// BleedingWoods is the card definition for Bleeding Woods.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless a player has 13 or less life.
//	{T}: Add {R} or {G}.
var BleedingWoods = newBleedingWoods

func newBleedingWoods() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:  "Bleeding Woods",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless a player has 13 or less life.", &game.Condition{
					Negate:              true,
					AnyPlayerLifeAtMost: 13,
				}),
			},
			OracleText: `
			This land enters tapped unless a player has 13 or less life.
			{T}: Add {R} or {G}.
		`,
		},
	}
}
