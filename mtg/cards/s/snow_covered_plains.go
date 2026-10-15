package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SnowCoveredPlains is the card definition for Snow-Covered Plains.
//
// Type: Basic Snow Land — Plains
//
// Oracle text:
//
//	({T}: Add {W}.)
var SnowCoveredPlains = newSnowCoveredPlains

func newSnowCoveredPlains() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Snow-Covered Plains",
			Supertypes: []types.Super{types.Basic, types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			OracleText: `
			({T}: Add {W}.)
		`,
		},
	}
}
