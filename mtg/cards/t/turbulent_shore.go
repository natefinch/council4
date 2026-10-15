package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurbulentShore is the card definition for Turbulent Shore.
//
// Type: Land — Plains Island
//
// Oracle text:
//
//	({T}: Add {W} or {U}.)
//	This land enters tapped unless your opponents control eight or more lands.
var TurbulentShore = newTurbulentShore

func newTurbulentShore() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name:     "Turbulent Shore",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Plains, types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.W, mana.U),
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
			({T}: Add {W} or {U}.)
			This land enters tapped unless your opponents control eight or more lands.
		`,
		},
	}
}
