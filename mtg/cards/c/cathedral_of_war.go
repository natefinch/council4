package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// CathedralOfWar is the card definition for Cathedral of War.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
//	{T}: Add {C}.
var CathedralOfWar = newCathedralOfWar

func newCathedralOfWar() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Cathedral of War",
			Types: []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.ExaltedStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			Exalted (Whenever a creature you control attacks alone, that creature gets +1/+1 until end of turn.)
			{T}: Add {C}.
		`,
		},
	}
}
