package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GoldmireBridge is the card definition for Goldmire Bridge.
//
// Type: Artifact Land
//
// Oracle text:
//
//	This land enters tapped.
//	Indestructible
//	{T}: Add {W} or {B}.
var GoldmireBridge = newGoldmireBridge

func newGoldmireBridge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:  "Goldmire Bridge",
			Types: []types.Card{types.Artifact, types.Land},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			Indestructible
			{T}: Add {W} or {B}.
		`,
		},
	}
}
