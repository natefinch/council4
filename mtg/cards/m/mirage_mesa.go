package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// MirageMesa is the card definition for Mirage Mesa.
//
// Type: Land — Desert
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color.
//	{T}: Add one mana of the chosen color.
var MirageMesa = newMirageMesa

func newMirageMesa() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Mirage Mesa",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Desert},
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
