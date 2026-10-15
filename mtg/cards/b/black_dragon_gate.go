package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// BlackDragonGate is the card definition for Black Dragon Gate.
//
// Type: Land — Gate
//
// Oracle text:
//
//	This land enters tapped.
//	As this land enters, choose a color other than black.
//	{T}: Add {B} or one mana of the chosen color.
var BlackDragonGate = newBlackDragonGate

func newBlackDragonGate() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:     "Black Dragon Gate",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Gate},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {B} or one mana of the chosen color.", mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
				game.EntryColorChoiceExcludingReplacement("As this land enters, choose a color other than black.", mana.B),
			},
			OracleText: `
			This land enters tapped.
			As this land enters, choose a color other than black.
			{T}: Add {B} or one mana of the chosen color.
		`,
		},
	}
}
