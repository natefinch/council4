package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ThrivingGrove is the card definition for Thriving Grove.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color other than green.
//	{T}: Add {G} or one mana of the chosen color.
var ThrivingGrove = newThrivingGrove

func newThrivingGrove() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:  "Thriving Grove",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapFixedOrChosenColorManaAbility("{T}: Add {G} or one mana of the chosen color.", mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceExcludingReplacement("This land enters tapped. As it enters, choose a color other than green.", mana.G),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color other than green.
			{T}: Add {G} or one mana of the chosen color.
		`,
		},
	}
}
