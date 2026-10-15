package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// HengegatePathway is the card definition for Hengegate Pathway // Mistgate Pathway.
//
// Type: Land // Land
// Face: Mistgate Pathway — Land
//
// Oracle text:
//
//	{T}: Add {W}.
var HengegatePathway = newHengegatePathway

func newHengegatePathway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:  "Hengegate Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			OracleText: `
			{T}: Add {W}.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Mistgate Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			OracleText: `
			{T}: Add {U}.
		`,
		}),
	}
}
