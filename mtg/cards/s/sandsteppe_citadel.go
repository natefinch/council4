package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// SandsteppeCitadel is the card definition for Sandsteppe Citadel.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {W}, {B}, or {G}.
var SandsteppeCitadel = newSandsteppeCitadel

func newSandsteppeCitadel() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black, color.Green),
		CardFace: game.CardFace{
			Name:  "Sandsteppe Citadel",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {W}, {B}, or {G}.
		`,
		},
	}
}
