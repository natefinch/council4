package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Swamp is the card definition for Swamp.
//
// Type: Basic Land — Swamp
//
// Oracle text:
//
//	({T}: Add {B}.)
var Swamp = newSwamp

func newSwamp() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name:       "Swamp",
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			OracleText: `
			({T}: Add {B}.)
		`,
		},
	}
}
