package d

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// DedicatedCommons is the card definition for Dedicated Commons.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a planeswalker.
//	{T}: Add {R} or {W}.
var DedicatedCommons = newDedicatedCommons

func newDedicatedCommons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:  "Dedicated Commons",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
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
			{T}: Add {R} or {W}.
		`,
		},
	}
}
