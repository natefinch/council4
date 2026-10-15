package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// CrossroadsVillage is the card definition for Crossroads Village.
//
// Type: Land — Town
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color.
//	{T}: Add one mana of the chosen color.
var CrossroadsVillage = newCrossroadsVillage

func newCrossroadsVillage() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Crossroads Village",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceReplacement("This land enters tapped. As it enters, choose a color."),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color.
			{T}: Add one mana of the chosen color.
		`,
		},
	}
}
