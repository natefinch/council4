package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// MoltenTributary is the card definition for Molten Tributary.
//
// Type: Land — Island Mountain
//
// Oracle text:
//
//	({T}: Add {U} or {R}.)
//	This land enters tapped.
var MoltenTributary = newMoltenTributary

func newMoltenTributary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Molten Tributary",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U} or {R}.)
			This land enters tapped.
		`,
		},
	}
}
