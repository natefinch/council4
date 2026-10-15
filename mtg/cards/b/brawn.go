package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Brawn is the card definition for Brawn.
//
// Type: Creature — Incarnation
// Cost: {3}{G}
//
// Oracle text:
//
//	Trample
//	As long as this card is in your graveyard and you control a Forest, creatures you control have trample.
var Brawn = newBrawn

func newBrawn() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Brawn",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Incarnation},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.StaticAbility{
					Text: "As long as this card is in your graveyard and you control a Forest, creatures you control have trample.",
					Condition: opt.Val(game.Condition{
						Text: "As long as this card is in your graveyard and you control a Forest",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Forest}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ZoneOfFunction: zone.Graveyard,
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration:    game.DurationPermanent,
							Group:       game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:       game.LayerAbility,
							AddKeywords: []game.Keyword{game.Trample},
						},
					},
				},
			},
			OracleText: `
			Trample
			As long as this card is in your graveyard and you control a Forest, creatures you control have trample.
		`,
		},
	}
}
