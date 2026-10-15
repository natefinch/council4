package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RadiantSummit is the card definition for Radiant Summit.
//
// Type: Land — Mountain Plains
//
// Oracle text:
//
//	({T}: Add {R} or {W}.)
//	This land enters tapped unless you control two or more basic lands.
var RadiantSummit = newRadiantSummit

func newRadiantSummit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:     "Radiant Summit",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control two or more basic lands.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
						MinCount:  2,
					}),
				}),
			},
			OracleText: `
			({T}: Add {R} or {W}.)
			This land enters tapped unless you control two or more basic lands.
		`,
		},
	}
}
