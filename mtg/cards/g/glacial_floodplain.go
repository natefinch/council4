package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GlacialFloodplain is the card definition for Glacial Floodplain.
//
// Type: Snow Land — Plains Island
//
// Oracle text:
//
//	({T}: Add {W} or {U}.)
//	This land enters tapped.
var GlacialFloodplain = newGlacialFloodplain

func newGlacialFloodplain() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:       "Glacial Floodplain",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Plains, types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {W} or {U}.)
			This land enters tapped.
		`,
		},
	}
}
