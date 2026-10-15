package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurbulentSprings is the card definition for Turbulent Springs.
//
// Type: Land — Island Mountain
//
// Oracle text:
//
//	({T}: Add {U} or {R}.)
//	This land enters tapped unless your opponents control eight or more lands.
var TurbulentSprings = newTurbulentSprings

func newTurbulentSprings() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Red),
		CardFace: game.CardFace{
			Name:     "Turbulent Springs",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Island, types.Mountain},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.U, mana.R),
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
			({T}: Add {U} or {R}.)
			This land enters tapped unless your opponents control eight or more lands.
		`,
		},
	}
}
