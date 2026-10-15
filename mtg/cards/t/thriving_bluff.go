package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ThrivingBluff is the card definition for Thriving Bluff.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color other than red.
//	{T}: Add {R} or one mana of the chosen color.
var ThrivingBluff = newThrivingBluff

func newThrivingBluff() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Thriving Bluff",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {R} or one mana of the chosen color.", mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceExcludingReplacement("This land enters tapped. As it enters, choose a color other than red.", mana.R),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color other than red.
			{T}: Add {R} or one mana of the chosen color.
		`,
		},
	}
}
