package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ThrivingIsle is the card definition for Thriving Isle.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color other than blue.
//	{T}: Add {U} or one mana of the chosen color.
var ThrivingIsle = newThrivingIsle

func newThrivingIsle() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Thriving Isle",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {U} or one mana of the chosen color.", mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceExcludingReplacement("This land enters tapped. As it enters, choose a color other than blue.", mana.U),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color other than blue.
			{T}: Add {U} or one mana of the chosen color.
		`,
		},
	}
}
