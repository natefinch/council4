package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DarksteelIngot is the card definition for Darksteel Ingot.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Indestructible (Effects that say "destroy" don't destroy this artifact.)
//	{T}: Add one mana of any color.
var DarksteelIngot = newDarksteelIngot

func newDarksteelIngot() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Darksteel Ingot",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.IndestructibleStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			Indestructible (Effects that say "destroy" don't destroy this artifact.)
			{T}: Add one mana of any color.
		`,
		},
	}
}
