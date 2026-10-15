package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurbulentWilderness is the card definition for Turbulent Wilderness.
//
// Type: Land — Forest Island
//
// Oracle text:
//
//	({T}: Add {G} or {U}.)
//	This land enters tapped unless your opponents control eight or more lands.
var TurbulentWilderness = newTurbulentWilderness

func newTurbulentWilderness() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name:     "Turbulent Wilderness",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Forest, types.Island},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.G, mana.U),
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
			({T}: Add {G} or {U}.)
			This land enters tapped unless your opponents control eight or more lands.
		`,
		},
	}
}
