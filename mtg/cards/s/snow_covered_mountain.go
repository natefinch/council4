package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SnowCoveredMountain is the card definition for Snow-Covered Mountain.
//
// Type: Basic Snow Land — Mountain
//
// Oracle text:
//
//	({T}: Add {R}.)
var SnowCoveredMountain = newSnowCoveredMountain

func newSnowCoveredMountain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:       "Snow-Covered Mountain",
			Supertypes: []types.Super{types.Basic, types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			OracleText: `
			({T}: Add {R}.)
		`,
		},
	}
}
