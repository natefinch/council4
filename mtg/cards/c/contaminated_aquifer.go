package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ContaminatedAquifer is the card definition for Contaminated Aquifer.
//
// Type: Land — Island Swamp
//
// Oracle text:
//
//	({T}: Add {U} or {B}.)
//	This land enters tapped.
var ContaminatedAquifer = newContaminatedAquifer

func newContaminatedAquifer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:     "Contaminated Aquifer",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {U} or {B}.)
			This land enters tapped.
		`,
		},
	}
}
