package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Replicated Ring
//
// Type: Token Snow Artifact
//
// Oracle text:
//   {T}: Add one mana of any color.

// ReplicatedRingToken23ee9957a9b74c2e8bf695e89a0dac7d is the card definition for Replicated Ring.
var ReplicatedRingToken23ee9957a9b74c2e8bf695e89a0dac7d = &game.CardDef{
	CardFace: game.CardFace{
		Name:       "Replicated Ring",
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
