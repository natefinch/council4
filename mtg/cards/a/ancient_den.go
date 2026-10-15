package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// AncientDen is the card definition for Ancient Den.
//
// Type: Artifact Land
//
// Oracle text:
//
//	{T}: Add {W}.
var AncientDen = newAncientDen

func newAncientDen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name:  "Ancient Den",
			Types: []types.Card{types.Artifact, types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			OracleText: `
			{T}: Add {W}.
		`,
		},
	}
}
