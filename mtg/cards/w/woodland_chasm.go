package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// WoodlandChasm is the card definition for Woodland Chasm.
//
// Type: Snow Land — Swamp Forest
//
// Oracle text:
//
//	({T}: Add {B} or {G}.)
//	This land enters tapped.
var WoodlandChasm = newWoodlandChasm

func newWoodlandChasm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:       "Woodland Chasm",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Swamp, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {B} or {G}.)
			This land enters tapped.
		`,
		},
	}
}
