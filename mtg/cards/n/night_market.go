package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
)

// NightMarket is the card definition for Night Market.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped. As it enters, choose a color.
//	{T}: Add one mana of the chosen color.
//	Cycling {3} ({3}, Discard this card: Draw a card.)
var NightMarket = newNightMarket

func newNightMarket() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Night Market",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(3)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapChosenColorManaAbility("{T}: Add one mana of the chosen color."),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedColorChoiceReplacement("This land enters tapped. As it enters, choose a color."),
			},
			OracleText: `
			This land enters tapped. As it enters, choose a color.
			{T}: Add one mana of the chosen color.
			Cycling {3} ({3}, Discard this card: Draw a card.)
		`,
		},
	}
}
