package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// FrostMarsh is the card definition for Frost Marsh.
//
// Type: Snow Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U} or {B}.
var FrostMarsh = newFrostMarsh

func newFrostMarsh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:       "Frost Marsh",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {U} or {B}.
		`,
		},
	}
}
