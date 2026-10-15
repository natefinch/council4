package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// ReliquaryTower is the card definition for Reliquary Tower.
//
// Type: Land
//
// Oracle text:
//
//	You have no maximum hand size.
//	{T}: Add {C}.
var ReliquaryTower = newReliquaryTower

func newReliquaryTower() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Reliquary Tower",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			You have no maximum hand size.
			{T}: Add {C}.
		`,
		},
	}
}
