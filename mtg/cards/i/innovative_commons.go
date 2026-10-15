package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// InnovativeCommons is the card definition for Innovative Commons.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a planeswalker.
//	{T}: Add {U} or {R}.
var InnovativeCommons = newInnovativeCommons

func newInnovativeCommons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:  "Innovative Commons",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
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
			{T}: Add {U} or {R}.
		`,
		},
	}
}
