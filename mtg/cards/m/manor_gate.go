package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ManorGate is the card definition for Manor Gate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	As this land enters, choose a color other than green.
//	{T}: Add {G} or one mana of the chosen color.
var ManorGate = newManorGate

func newManorGate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:     "Manor Gate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {G} or one mana of the chosen color.", mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntryColorChoiceExcludingReplacement("As this land enters, choose a color other than green.", mana.G),
			},
			OracleText: `
			This land enters tapped.
			As this land enters, choose a color other than green.
			{T}: Add {G} or one mana of the chosen color.
		`,
		},
	}
}
