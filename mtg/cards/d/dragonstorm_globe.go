package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DragonstormGlobe is the card definition for Dragonstorm Globe.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Each Dragon you control enters with an additional +1/+1 counter on it.
//	{T}: Add one mana of any color.
var DragonstormGlobe = newDragonstormGlobe

func newDragonstormGlobe() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Dragonstorm Globe",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U, mana.B, mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersGroupReplacement("Each Dragon you control enters with an additional +1/+1 counter on it.", &game.Selection{SubtypesAny: []types.Sub{types.Sub("Dragon")}, Controller: game.ControllerYou}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Each Dragon you control enters with an additional +1/+1 counter on it.
			{T}: Add one mana of any color.
		`,
		},
	}
}
