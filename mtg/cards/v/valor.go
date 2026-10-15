package v

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Valor is the card definition for Valor.
//
// Type: Creature — Incarnation
// Cost: {3}{W}
//
// Oracle text:
//
//	First strike
//	As long as this card is in your graveyard and you control a Plains, creatures you control have first strike.
var Valor = newValor

func newValor() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Valor",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Incarnation},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
				game.StaticAbility{
					Text: "As long as this card is in your graveyard and you control a Plains, creatures you control have first strike.",
					Condition: opt.Val(game.Condition{
						Text: "As long as this card is in your graveyard and you control a Plains",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Plains}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ZoneOfFunction: zone.Graveyard,
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.FirstStrike},
						},
					},
				},
			},
			OracleText: `
			First strike
			As long as this card is in your graveyard and you control a Plains, creatures you control have first strike.
		`,
		},
	}
}
