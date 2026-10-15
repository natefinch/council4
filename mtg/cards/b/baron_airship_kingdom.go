package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// BaronAirshipKingdom is the card definition for Baron, Airship Kingdom.
//
// Type: Land — Town
//
// Oracle text:
//
//	This land enters tapped.
//	{T}: Add {U} or {R}.
var BaronAirshipKingdom = newBaronAirshipKingdom

func newBaronAirshipKingdom() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Baron, Airship Kingdom",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Town},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			{T}: Add {U} or {R}.
		`,
		},
	}
}
