package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// WoodedRidgeline is the card definition for Wooded Ridgeline.
//
// Type: Land — Mountain Forest
//
// Oracle text:
//
//	({T}: Add {R} or {G}.)
//	This land enters tapped.
var WoodedRidgeline = newWoodedRidgeline

func newWoodedRidgeline() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:     "Wooded Ridgeline",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {R} or {G}.)
			This land enters tapped.
		`,
		},
	}
}
