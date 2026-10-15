package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Island is the card definition for Island.
//
// Type: Basic Land — Island
//
// Oracle text:
//
//	({T}: Add {U}.)
var Island = newIsland

func newIsland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:       "Island",
			Supertypes: []types.Super{types.Basic},
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
