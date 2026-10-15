package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SnowCoveredSwamp is the card definition for Snow-Covered Swamp.
//
// Type: Basic Snow Land — Swamp
//
// Oracle text:
//
//	({T}: Add {B}.)
var SnowCoveredSwamp = newSnowCoveredSwamp

func newSnowCoveredSwamp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Snow-Covered Swamp",
			Supertypes: []types.Super{types.Basic, types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			OracleText: `
			({T}: Add {B}.)
		`,
		},
	}
}
