package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// AlpineMeadow is the card definition for Alpine Meadow.
//
// Type: Snow Land — Mountain Plains
//
// Oracle text:
//
//	({T}: Add {R} or {W}.)
//	This land enters tapped.
var AlpineMeadow = newAlpineMeadow

func newAlpineMeadow() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:       "Alpine Meadow",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Mountain, types.Plains},
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
