package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SnowCoveredForest is the card definition for Snow-Covered Forest.
//
// Type: Basic Snow Land — Forest
//
// Oracle text:
//
//	({T}: Add {G}.)
var SnowCoveredForest = newSnowCoveredForest

func newSnowCoveredForest() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name:       "Snow-Covered Forest",
			Supertypes: []types.Super{types.Basic, types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			OracleText: `
			({T}: Add {G}.)
		`,
		},
	}
}
