package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// UnchartedHaven is the card definition for Uncharted Haven.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color.
//	{T}: Add one mana of the chosen color.
var UnchartedHaven = newUnchartedHaven

func newUnchartedHaven() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Uncharted Haven",
			Types: []types.Card{types.Land},
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
