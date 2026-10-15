package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CragcrownPathway is the card definition for Cragcrown Pathway // Timbercrown Pathway.
//
// Type: Land // Land
// Face: Timbercrown Pathway — Land
//
// Oracle text:
//
//	{T}: Add {R}.
var CragcrownPathway = newCragcrownPathway

func newCragcrownPathway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:  "Cragcrown Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.R),
			},
			OracleText: `
			{T}: Add {R}.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Timbercrown Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.G),
			},
			OracleText: `
			{T}: Add {G}.
		`,
		}),
	}
}
