package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ThrivingHeath is the card definition for Thriving Heath.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color other than white.
//	{T}: Add {W} or one mana of the chosen color.
var ThrivingHeath = newThrivingHeath

func newThrivingHeath() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Thriving Heath",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {W} or one mana of the chosen color.", mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceExcludingReplacement("This land enters tapped. As it enters, choose a color other than white.", mana.W),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color other than white.
			{T}: Add {W} or one mana of the chosen color.
		`,
		},
	}
}
