package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TurbulentFen is the card definition for Turbulent Fen.
//
// Type: Land — Swamp Forest
//
// Oracle text:
//
//	({T}: Add {B} or {G}.)
//	This land enters tapped unless your opponents control eight or more lands.
var TurbulentFen = newTurbulentFen

func newTurbulentFen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Green),
		CardFace: game.CardFace{
			Name:     "Turbulent Fen",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Swamp, types.Forest},
			ManaAbilities: []game.ManaAbility{
				game.TapManaChoiceAbility(mana.B, mana.G),
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
			({T}: Add {B} or {G}.)
			This land enters tapped unless your opponents control eight or more lands.
		`,
		},
	}
}
