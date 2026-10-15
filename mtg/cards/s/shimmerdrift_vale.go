package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// ShimmerdriftVale is the card definition for Shimmerdrift Vale.
//
// Type: Snow Land
//
// Oracle text:
//
//	This land enters tapped.
//	As this land enters, choose a color.
//	{T}: Add one mana of the chosen color.
var ShimmerdriftVale = newShimmerdriftVale

func newShimmerdriftVale() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Shimmerdrift Vale",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntryColorChoiceReplacement("As this land enters, choose a color."),
			},
			OracleText: `
			This land enters tapped.
			As this land enters, choose a color.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
