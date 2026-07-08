package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TeethingWurmlet is the card definition for Teething Wurmlet.
//
// Type: Creature — Wurm
// Cost: {G}
//
// Oracle text:
//
//	This creature has deathtouch as long as you control three or more artifacts.
//	Whenever an artifact you control enters, you gain 1 life. If this is the first time this ability has resolved this turn, put a +1/+1 counter on this creature.
var TeethingWurmlet = newTeethingWurmlet

func newTeethingWurmlet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Teething Wurmlet",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Wurm},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						ControlsMatching: opt.Val(game.SelectionCount{
							Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
							MinCount:  3,
						}),
					}),
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer:          game.LayerAbility,
							AffectedSource: true,
							AddKeywords: []game.Keyword{
								game.Deathtouch,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
						},
					},
					CountsResolutionsThisTurn: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										SourceAbilityResolutionOrdinalThisTurn: 1,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This creature has deathtouch as long as you control three or more artifacts.
			Whenever an artifact you control enters, you gain 1 life. If this is the first time this ability has resolved this turn, put a +1/+1 counter on this creature.
		`,
		},
	}
}
