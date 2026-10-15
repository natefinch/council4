package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SnowCoveredIsland is the card definition for Snow-Covered Island.
//
// Type: Basic Snow Land — Island
//
// Oracle text:
//
//	({T}: Add {U}.)
var SnowCoveredIsland = newSnowCoveredIsland

func newSnowCoveredIsland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Snow-Covered Island",
			Supertypes: []types.Super{types.Basic, types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			OracleText: `
			({T}: Add {U}.)
		`,
		},
	}
}
