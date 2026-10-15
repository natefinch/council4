package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TransformativeCommons is the card definition for Transformative Commons.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a planeswalker.
//	{T}: Add {G} or {U}.
var TransformativeCommons = newTransformativeCommons

func newTransformativeCommons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:  "Transformative Commons",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a planeswalker.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a planeswalker.
			{T}: Add {G} or {U}.
		`,
		},
	}
}
