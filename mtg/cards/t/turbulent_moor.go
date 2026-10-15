package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurbulentMoor is the card definition for Turbulent Moor.
//
// Type: Land — Plains Swamp
//
// Oracle text:
//
//	({T}: Add {W} or {B}.)
//	This land enters tapped unless your opponents control eight or more lands.
var TurbulentMoor = newTurbulentMoor

func newTurbulentMoor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Black),
		CardFace: game.CardFace{
			Name:     "Turbulent Moor",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Swamp},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.B),
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedIfReplacement("This land enters tapped unless your opponents control eight or more lands.", &game.Condition{
					Negate: true,
					OpponentsControl: opt.Val(game.SelectionCount{
						Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
						MinCount:  8,
					}),
				}),
			},
			OracleText: `
			({T}: Add {W} or {B}.)
			This land enters tapped unless your opponents control eight or more lands.
		`,
		},
	}
}
