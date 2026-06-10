package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Icy Manalith
//
// Type: Token Snow Artifact
//
// Oracle text:
//   {T}: Add one mana of any color.

// IcyManalithToken2944167b0ffe4582b14db6f80d936d65 is the card definition for Icy Manalith.
var IcyManalithToken2944167b0ffe4582b14db6f80d936d65 = &game.CardDef{
	CardFace: game.CardFace{
		Name:       "Icy Manalith",
		Supertypes: []types.Super{types.Snow},
		Types:      []types.Card{types.Artifact},
		ManaAbilities: []game.ManaAbility{
			game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
		},
		OracleText: `
			{T}: Add one mana of any color.
		`,
	},
}
