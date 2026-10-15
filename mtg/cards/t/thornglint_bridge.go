package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ThornglintBridge is the card definition for Thornglint Bridge.
//
// Type: Artifact Land
//
// Oracle text:
//
//	This land enters tapped.
//	Indestructible
//	{T}: Add {G} or {W}.
var ThornglintBridge = newThornglintBridge

func newThornglintBridge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name:  "Thornglint Bridge",
			Types: []types.Card{types.Artifact, types.Land},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			Indestructible
			{T}: Add {G} or {W}.
		`,
		},
	}
}
