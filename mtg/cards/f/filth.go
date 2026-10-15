package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// Filth is the card definition for Filth.
//
// Type: Creature — Incarnation
// Cost: {3}{B}
//
// Oracle text:
//
//	Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
//	As long as this card is in your graveyard and you control a Swamp, creatures you control have swampwalk.
var Filth = newFilth

func newFilth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Filth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Incarnation},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.SwampwalkStaticBody,
				game.StaticAbility{
					Text: "As long as this card is in your graveyard and you control a Swamp, creatures you control have swampwalk.",
					Condition: opt.Val(game.Condition{
						Text: "As long as this card is in your graveyard and you control a Swamp",
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{SubtypesAny: []types.Sub{types.Swamp}},
						}),
						SourceCounterKind: counter.PlusOnePlusOne,
					}),
					ZoneOfFunction: zone.Graveyard,
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Duration: game.DurationPermanent,
							Group:    game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							Layer:    game.LayerAbility,
							AddAbilities: []game.Ability{
								&game.StaticAbility{
									Text: "Swampwalk",
									KeywordAbilities: []game.KeywordAbility{
										game.LandwalkKeyword{
											Subtype: types.Swamp,
										},
									},
								},
							},
						},
					},
				},
			},
			OracleText: `
			Swampwalk (This creature can't be blocked as long as defending player controls a Swamp.)
			As long as this card is in your graveyard and you control a Swamp, creatures you control have swampwalk.
		`,
		},
	}
}
