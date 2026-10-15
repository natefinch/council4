package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ObeliskOfJund is the card definition for Obelisk of Jund.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {B}, {R}, or {G}.
var ObeliskOfJund = newObeliskOfJund

func newObeliskOfJund() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Obelisk of Jund",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R, mana.G),
			},
			OracleText: `
			{T}: Add {B}, {R}, or {G}.
		`,
		},
	}
}
