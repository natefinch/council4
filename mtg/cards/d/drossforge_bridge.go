package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DrossforgeBridge is the card definition for Drossforge Bridge.
//
// Type: Artifact Land
//
// Oracle text:
//
//	This land enters tapped.
//	Indestructible
//	{T}: Add {B} or {R}.
var DrossforgeBridge = newDrossforgeBridge

func newDrossforgeBridge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:  "Drossforge Bridge",
			Types: []types.Card{types.Artifact, types.Land},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			Indestructible
			{T}: Add {B} or {R}.
		`,
		},
	}
}
