package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// VolcanicIsland is the card definition for Volcanic Island.
//
// Type: Land — Island Mountain
//
// Oracle text:
//
//	({T}: Add {U} or {R}.)
var VolcanicIsland = newVolcanicIsland

func newVolcanicIsland() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Volcanic Island",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			OracleText: `
			({T}: Add {U} or {R}.)
		`,
		},
	}
}
