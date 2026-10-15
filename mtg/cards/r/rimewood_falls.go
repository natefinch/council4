package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RimewoodFalls is the card definition for Rimewood Falls.
//
// Type: Snow Land — Forest Island
//
// Oracle text:
//
//	({T}: Add {G} or {U}.)
//	This land enters tapped.
var RimewoodFalls = newRimewoodFalls

func newRimewoodFalls() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:       "Rimewood Falls",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Forest, types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G} or {U}.)
			This land enters tapped.
		`,
		},
	}
}
