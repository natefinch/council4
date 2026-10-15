package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Wonder is the card definition for Wonder.
//
// Type: Creature — Incarnation
// Cost: {3}{U}
//
// Oracle text:
//
//	Flying
//	As long as this card is in your graveyard and you control an Island, creatures you control have flying.
var Wonder = newWonder

func newWonder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Wonder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Incarnation},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.StaticAbility{
					Text: "As long as this card is in your graveyard and you control an Island, creatures you control have flying.",
					Condition: opt.Val(game.Condition{
						Text: "As long as this card is in your graveyard and you control an Island",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Island}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ZoneOfFunction: zone.Graveyard,
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Flying},
						},
					},
				},
			},
			OracleText: `
			Flying
			As long as this card is in your graveyard and you control an Island, creatures you control have flying.
		`,
		},
	}
}
