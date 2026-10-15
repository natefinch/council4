package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Wastes is the card definition for Wastes.
//
// Type: Basic Land
//
// Oracle text:
//
//	{T}: Add {C}.
var Wastes = newWastes

func newWastes() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Wastes",
			Supertypes: []types.Super{types.Basic},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
		`,
		},
	}
}
