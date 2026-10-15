package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/types"
)

// ReflectingPool is the card definition for Reflecting Pool.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add one mana of any type that a land you control could produce.
var ReflectingPool = newReflectingPool

func newReflectingPool() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Reflecting Pool",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaLandsProduceAbility(game.PlayerYou, true),
			},
			OracleText: `
			{T}: Add one mana of any type that a land you control could produce.
		`,
		},
	}
}
