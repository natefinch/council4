package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SulfurousMire is the card definition for Sulfurous Mire.
//
// Type: Snow Land — Swamp Mountain
//
// Oracle text:
//
//	({T}: Add {B} or {R}.)
//	This land enters tapped.
var SulfurousMire = newSulfurousMire

func newSulfurousMire() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:       "Sulfurous Mire",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Swamp, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {B} or {R}.)
			This land enters tapped.
		`,
		},
	}
}
