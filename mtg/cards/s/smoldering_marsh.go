package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SmolderingMarsh is the card definition for Smoldering Marsh.
//
// Type: Land — Swamp Mountain
//
// Oracle text:
//
//	({T}: Add {B} or {R}.)
//	This land enters tapped unless you control two or more basic lands.
var SmolderingMarsh = newSmolderingMarsh

func newSmolderingMarsh() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name:     "Smoldering Marsh",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.R),
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
			({T}: Add {B} or {R}.)
			This land enters tapped unless you control two or more basic lands.
		`,
		},
	}
}
