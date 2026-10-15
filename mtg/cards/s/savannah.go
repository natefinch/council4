package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Savannah is the card definition for Savannah.
//
// Type: Land — Forest Plains
//
// Oracle text:
//
//	({T}: Add {G} or {W}.)
var Savannah = newSavannah

func newSavannah() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:     "Savannah",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			OracleText: `
			({T}: Add {G} or {W}.)
		`,
		},
	}
}
