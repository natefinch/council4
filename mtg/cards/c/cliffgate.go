package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Cliffgate is the card definition for Cliffgate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	As this land enters, choose a color other than red.
//	{T}: Add {R} or one mana of the chosen color.
var Cliffgate = newCliffgate

func newCliffgate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:     "Cliffgate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {R} or one mana of the chosen color.", mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntryColorChoiceExcludingReplacement("As this land enters, choose a color other than red.", mana.R),
			},
			OracleText: `
			This land enters tapped.
			As this land enters, choose a color other than red.
			{T}: Add {R} or one mana of the chosen color.
		`,
		},
	}
}
