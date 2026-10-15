package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// GreatFurnace is the card definition for Great Furnace.
//
// Type: Artifact Land
//
// Oracle text:
//
//	{T}: Add {R}.
var GreatFurnace = newGreatFurnace

func newGreatFurnace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name:  "Great Furnace",
			Types: []types.Card{types.Artifact, types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			OracleText: `
			{T}: Add {R}.
		`,
		},
	}
}
