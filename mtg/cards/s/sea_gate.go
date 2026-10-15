package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SeaGate is the card definition for Sea Gate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	As this land enters, choose a color other than blue.
//	{T}: Add {U} or one mana of the chosen color.
var SeaGate = newSeaGate

func newSeaGate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:     "Sea Gate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {U} or one mana of the chosen color.", mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntryColorChoiceExcludingReplacement("As this land enters, choose a color other than blue.", mana.U),
			},
			OracleText: `
			This land enters tapped.
			As this land enters, choose a color other than blue.
			{T}: Add {U} or one mana of the chosen color.
		`,
		},
	}
}
