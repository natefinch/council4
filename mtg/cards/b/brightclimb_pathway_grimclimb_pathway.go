package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrightclimbPathway is the card definition for Brightclimb Pathway // Grimclimb Pathway.
//
// Type: Land // Land
// Face: Grimclimb Pathway — Land
//
// Oracle text:
//
//	{T}: Add {W}.
var BrightclimbPathway = newBrightclimbPathway

func newBrightclimbPathway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:  "Brightclimb Pathway",
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
			Name:  "Grimclimb Pathway",
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
