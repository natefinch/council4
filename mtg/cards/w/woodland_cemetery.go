package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WoodlandCemetery is the card definition for Woodland Cemetery.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Swamp or a Forest.
//	{T}: Add {B} or {G}.
var WoodlandCemetery = newWoodlandCemetery

func newWoodlandCemetery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:  "Woodland Cemetery",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Swamp or a Forest.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Swamp"), types.Sub("Forest")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Swamp or a Forest.
			{T}: Add {B} or {G}.
		`,
		},
	}
}
