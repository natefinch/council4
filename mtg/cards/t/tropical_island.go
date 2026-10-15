package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TropicalIsland is the card definition for Tropical Island.
//
// Type: Land — Forest Island
//
// Oracle text:
//
//	({T}: Add {G} or {U}.)
var TropicalIsland = newTropicalIsland

func newTropicalIsland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Tropical Island",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			OracleText: `
			({T}: Add {G} or {U}.)
		`,
		},
	}
}
