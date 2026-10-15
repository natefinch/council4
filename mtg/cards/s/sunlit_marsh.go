package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SunlitMarsh is the card definition for Sunlit Marsh.
//
// Type: Land — Plains Swamp
//
// Oracle text:
//
//	({T}: Add {W} or {B}.)
//	This land enters tapped.
var SunlitMarsh = newSunlitMarsh

func newSunlitMarsh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Sunlit Marsh",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {W} or {B}.)
			This land enters tapped.
		`,
		},
	}
}
