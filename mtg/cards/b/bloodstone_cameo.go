package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BloodstoneCameo is the card definition for Bloodstone Cameo.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	{T}: Add {B} or {R}.
var BloodstoneCameo = newBloodstoneCameo

func newBloodstoneCameo() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Bloodstone Cameo",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
			},
			OracleText: `
			{T}: Add {B} or {R}.
		`,
		},
	}
}
