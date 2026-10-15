package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// PeculiarLighthouse is the card definition for Peculiar Lighthouse.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless a player has 13 or less life.
//	{T}: Add {U} or {R}.
var PeculiarLighthouse = newPeculiarLighthouse

func newPeculiarLighthouse() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:  "Peculiar Lighthouse",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless a player has 13 or less life.", &game.Condition{
					Negate:              true,
					AnyPlayerLifeAtMost: 13,
				}),
			},
			OracleText: `
			This land enters tapped unless a player has 13 or less life.
			{T}: Add {U} or {R}.
		`,
		},
	}
}
