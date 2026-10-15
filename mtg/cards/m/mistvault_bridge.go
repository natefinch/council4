package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// MistvaultBridge is the card definition for Mistvault Bridge.
//
// Type: Artifact Land
//
// Oracle text:
//
//	This land enters tapped.
//	Indestructible
//	{T}: Add {U} or {B}.
var MistvaultBridge = newMistvaultBridge

func newMistvaultBridge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:  "Mistvault Bridge",
			Types: []types.Card{types.Artifact, types.Land},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			Indestructible
			{T}: Add {U} or {B}.
		`,
		},
	}
}
