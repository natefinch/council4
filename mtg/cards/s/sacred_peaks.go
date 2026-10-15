package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SacredPeaks is the card definition for Sacred Peaks.
//
// Type: Land — Mountain Plains
//
// Oracle text:
//
//	({T}: Add {R} or {W}.)
//	This land enters tapped.
var SacredPeaks = newSacredPeaks

func newSacredPeaks() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:     "Sacred Peaks",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {R} or {W}.)
			This land enters tapped.
		`,
		},
	}
}
