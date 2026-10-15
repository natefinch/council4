package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ArcticFlats is the card definition for Arctic Flats.
//
// Type: Snow Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {G} or {W}.
var ArcticFlats = newArcticFlats

func newArcticFlats() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:       "Arctic Flats",
			Supertypes: []types.Super{types.Snow},
			Types:      []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {G} or {W}.
		`,
		},
	}
}
