package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DecanterOfEndlessWater is the card definition for Decanter of Endless Water.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	You have no maximum hand size.
//	{T}: Add one mana of any color.
var DecanterOfEndlessWater = newDecanterOfEndlessWater

func newDecanterOfEndlessWater() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Decanter of Endless Water",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			OracleText: `
			You have no maximum hand size.
			{T}: Add one mana of any color.
		`,
		},
	}
}
