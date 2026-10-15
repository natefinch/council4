package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ClearwaterPathway is the card definition for Clearwater Pathway // Murkwater Pathway.
//
// Type: Land // Land
// Face: Murkwater Pathway — Land
//
// Oracle text:
//
//	{T}: Add {U}.
var ClearwaterPathway = newClearwaterPathway

func newClearwaterPathway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name:  "Clearwater Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.U),
			},
			OracleText: `
			{T}: Add {U}.
		`,
		},
		Layout: game.LayoutModalDFC,
		Back: opt.Val(game.CardFace{
			Name:  "Murkwater Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.B),
			},
			OracleText: `
			{T}: Add {B}.
		`,
		}),
	}
}
