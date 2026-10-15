package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ClifftopRetreat is the card definition for Clifftop Retreat.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped unless you control a Mountain or a Plains.
//	{T}: Add {R} or {W}.
var ClifftopRetreat = newClifftopRetreat

func newClifftopRetreat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:  "Clifftop Retreat",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless you control a Mountain or a Plains.", &game.Condition{
					Negate: true,
					ControlsMatching: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}, SubtypesAny: []types.Sub{types.Sub("Mountain"), types.Sub("Plains")}},
					}),
				}),
			},
			OracleText: `
			This land enters tapped unless you control a Mountain or a Plains.
			{T}: Add {R} or {W}.
		`,
		},
	}
}
