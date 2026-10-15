package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RootboundCrag is the card definition for Rootbound Crag.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Mountain or a Forest.
//	{T}: Add {R} or {G}.
var RootboundCrag = newRootboundCrag

func newRootboundCrag() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red, color.Green),
		CardFace: game.CardFace{
			Name:  "Rootbound Crag",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.G),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Mountain or a Forest.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Mountain"), types.Sub("Forest")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Mountain or a Forest.
			{T}: Add {R} or {G}.
		`,
		},
	}
}
