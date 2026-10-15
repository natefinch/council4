package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ArcticTreeline is the card definition for Arctic Treeline.
//
// Type: Snow Land — Forest Plains
//
// Oracle text:
//
//	({T}: Add {G} or {W}.)
//	This land enters tapped.
var ArcticTreeline = newArcticTreeline

func newArcticTreeline() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:       "Arctic Treeline",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			Subtypes:   []types.Sub{types.Forest, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			({T}: Add {G} or {W}.)
			This land enters tapped.
		`,
		},
	}
}
