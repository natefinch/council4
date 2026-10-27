// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// RTLand is the card definition for RT Land.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {G}.
var RTLand = newRTLand

func newRTLand() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Land",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			OracleText: `
			{T}: Add {G}.
		`,
		},
	}
}
