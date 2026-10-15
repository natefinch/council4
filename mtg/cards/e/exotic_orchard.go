package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// ExoticOrchard is the card definition for Exotic Orchard.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add one mana of any color that a land an opponent controls could produce.
var ExoticOrchard = newExoticOrchard

func newExoticOrchard() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Exotic Orchard",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaLandsProduceAbility(game.PlayerOpponent, false),
			},
			OracleText: `
			{T}: Add one mana of any color that a land an opponent controls could produce.
		`,
		},
	}
}
