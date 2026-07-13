package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TendershootDryad is the card definition for Tendershoot Dryad.
//
// Type: Creature — Dryad
// Cost: {4}{G}
//
// Oracle text:
//
//	Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
//	At the beginning of each upkeep, create a 1/1 green Saproling creature token.
//	Saprolings you control get +2/+2 as long as you have the city's blessing.
var TendershootDryad = newTendershootDryad

func newTendershootDryad() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Tendershoot Dryad",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Dryad},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.AscendStaticBody,
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControllerHasCityBlessing: true,
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerPowerToughnessModify,
							Group:          game.ObjectControlledGroup(game.SourcePermanentReference(), game.Selection{SubtypesAny: []types.Sub{types.Sub("Saproling")}}),
							PowerDelta:     2,
							ToughnessDelta: 2,
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(tendershootDryadToken),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Ascend (If you control ten or more permanents, you get the city's blessing for the rest of the game.)
			At the beginning of each upkeep, create a 1/1 green Saproling creature token.
			Saprolings you control get +2/+2 as long as you have the city's blessing.
		`,
		},
	}
}

var tendershootDryadToken = newTendershootDryadToken()

func newTendershootDryadToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Saproling",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Saproling},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
