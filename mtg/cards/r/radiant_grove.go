package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RadiantGrove is the card definition for Radiant Grove.
//
// Type: Land — Forest Plains
//
// Oracle text:
//
//	({T}: Add {G} or {W}.)
//	This land enters tapped.
var RadiantGrove = newRadiantGrove

func newRadiantGrove() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:     "Radiant Grove",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Plains},
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
