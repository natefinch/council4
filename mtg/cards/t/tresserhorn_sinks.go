package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TresserhornSinks is the card definition for Tresserhorn Sinks.
//
// Type: Snow Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {B} or {R}.
var TresserhornSinks = newTresserhornSinks

func newTresserhornSinks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:       "Tresserhorn Sinks",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {B} or {R}.
		`,
		},
	}
}
