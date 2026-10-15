package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Plains is the card definition for Plains.
//
// Type: Basic Land — Plains
//
// Oracle text:
//
//	({T}: Add {W}.)
var Plains = newPlains

func newPlains() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:       "Plains",
			Supertypes: []types.Super{types.Basic},
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
