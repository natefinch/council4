package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// CitadelGate is the card definition for Citadel Gate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	As this land enters, choose a color other than white.
//	{T}: Add {W} or one mana of the chosen color.
var CitadelGate = newCitadelGate

func newCitadelGate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:     "Citadel Gate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {W} or one mana of the chosen color.", mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntryColorChoiceExcludingReplacement("As this land enters, choose a color other than white.", mana.W),
			},
			OracleText: `
			This land enters tapped.
			As this land enters, choose a color other than white.
			{T}: Add {W} or one mana of the chosen color.
		`,
		},
	}
}
