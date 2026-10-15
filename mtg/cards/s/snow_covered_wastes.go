package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SnowCoveredWastes is the card definition for Snow-Covered Wastes.
//
// Type: Basic Snow Land
//
// Oracle text:
//
//	{T}: Add {C}.
var SnowCoveredWastes = newSnowCoveredWastes

func newSnowCoveredWastes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Snow-Covered Wastes",
			Supertypes: []types.Super{types.Basic, types.Snow},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
		`,
		},
	}
}
