package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObeliskOfGrixis is the card definition for Obelisk of Grixis.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {U}, {B}, or {R}.
var ObeliskOfGrixis = newObeliskOfGrixis

func newObeliskOfGrixis() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Obelisk of Grixis",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.B, mana.R),
			},
			OracleText: `
			{T}: Add {U}, {B}, or {R}.
		`,
		},
	}
}
