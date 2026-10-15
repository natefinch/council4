package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// BlastedLandscape is the card definition for Blasted Landscape.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	Cycling {2} ({2}, Discard this card: Draw a card.)
var BlastedLandscape = newBlastedLandscape

func newBlastedLandscape() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Blasted Landscape",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.CyclingActivatedAbility(cost.Mana{cost.O(2)}),
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			Cycling {2} ({2}, Discard this card: Draw a card.)
		`,
		},
	}
}
