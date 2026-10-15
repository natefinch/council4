package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Tundra is the card definition for Tundra.
//
// Type: Land — Plains Island
//
// Oracle text:
//
//	({T}: Add {W} or {U}.)
var Tundra = newTundra

func newTundra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:     "Tundra",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			OracleText: `
			({T}: Add {W} or {U}.)
		`,
		},
	}
}
