package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScorchedGeyser is the card definition for Scorched Geyser.
//
// Type: Land — Island Mountain
//
// Oracle text:
//
//	({T}: Add {U} or {R}.)
//	This land enters tapped unless you control two or more basic lands.
var ScorchedGeyser = newScorchedGeyser

func newScorchedGeyser() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Scorched Geyser",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
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
			({T}: Add {U} or {R}.)
			This land enters tapped unless you control two or more basic lands.
		`,
		},
	}
}
