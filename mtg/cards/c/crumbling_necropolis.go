package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// CrumblingNecropolis is the card definition for Crumbling Necropolis.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U}, {B}, or {R}.
var CrumblingNecropolis = newCrumblingNecropolis

func newCrumblingNecropolis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Crumbling Necropolis",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {U}, {B}, or {R}.
		`,
		},
	}
}
