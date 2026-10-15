package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// UndergroundSea is the card definition for Underground Sea.
//
// Type: Land — Island Swamp
//
// Oracle text:
//
//	({T}: Add {U} or {B}.)
var UndergroundSea = newUndergroundSea

func newUndergroundSea() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:     "Underground Sea",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			OracleText: `
			({T}: Add {U} or {B}.)
		`,
		},
	}
}
