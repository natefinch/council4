package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// BorealShelf is the card definition for Boreal Shelf.
//
// Type: Snow Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {W} or {U}.
var BorealShelf = newBorealShelf

func newBorealShelf() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:       "Boreal Shelf",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {W} or {U}.
		`,
		},
	}
}
