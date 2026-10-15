package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// DarksteelCitadel is the card definition for Darksteel Citadel.
//
// Type: Artifact Land
//
// Oracle text:
//
//	Indestructible
//	{T}: Add {C}.
var DarksteelCitadel = newDarksteelCitadel

func newDarksteelCitadel() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Darksteel Citadel",
			Types: []types.Card{types.Artifact, types.Land},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			Indestructible
			{T}: Add {C}.
		`,
		},
	}
}
