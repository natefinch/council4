package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// TangledIslet is the card definition for Tangled Islet.
//
// Type: Land — Forest Island
//
// Oracle text:
//
//	({T}: Add {G} or {U}.)
//	This land enters tapped.
var TangledIslet = newTangledIslet

func newTangledIslet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Tangled Islet",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Island},
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
