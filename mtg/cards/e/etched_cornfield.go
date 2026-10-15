package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// EtchedCornfield is the card definition for Etched Cornfield.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless a player has 13 or less life.
//	{T}: Add {G} or {W}.
var EtchedCornfield = newEtchedCornfield

func newEtchedCornfield() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:  "Etched Cornfield",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless a player has 13 or less life.", &game.Condition{
					Negate:              true,
					AnyPlayerLifeAtMost: 13,
				}),
			},
			OracleText: `
			This land enters tapped unless a player has 13 or less life.
			{T}: Add {G} or {W}.
		`,
		},
	}
}
