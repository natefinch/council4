package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NeedlevergePathway is the card definition for Needleverge Pathway // Pillarverge Pathway.
//
// Type: Land // Land
// Face: Pillarverge Pathway — Land
//
// Oracle text:
//
//	{T}: Add {R}.
var NeedlevergePathway = newNeedlevergePathway

func newNeedlevergePathway() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:  "Needleverge Pathway",
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
			Name:  "Pillarverge Pathway",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.W),
			},
			OracleText: `
			{T}: Add {W}.
		`,
		}),
	}
}
