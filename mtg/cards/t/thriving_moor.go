package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ThrivingMoor is the card definition for Thriving Moor.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color other than black.
//	{T}: Add {B} or one mana of the chosen color.
var ThrivingMoor = newThrivingMoor

func newThrivingMoor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:  "Thriving Moor",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {B} or one mana of the chosen color.", mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceExcludingReplacement("This land enters tapped. As it enters, choose a color other than black.", mana.B),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color other than black.
			{T}: Add {B} or one mana of the chosen color.
		`,
		},
	}
}
