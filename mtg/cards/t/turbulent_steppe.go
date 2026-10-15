package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurbulentSteppe is the card definition for Turbulent Steppe.
//
// Type: Land — Mountain Plains
//
// Oracle text:
//
//	({T}: Add {R} or {W}.)
//	This land enters tapped unless your opponents control eight or more lands.
var TurbulentSteppe = newTurbulentSteppe

func newTurbulentSteppe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name:     "Turbulent Steppe",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Mountain, types.Plains},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.R, mana.W),
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
			({T}: Add {R} or {W}.)
			This land enters tapped unless your opponents control eight or more lands.
		`,
		},
	}
}
