package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MsBumbleflower is the card definition for Ms. Bumbleflower.
//
// Type: Legendary Creature — Rabbit Citizen
// Cost: {1}{G}{W}{U}
//
// Oracle text:
//
//	Vigilance
//	Whenever you cast a spell, target opponent draws a card. Put a +1/+1 counter on target creature. It gains flying until end of turn. If this is the second time this ability has resolved this turn, you draw two cards.
var MsBumbleflower = newMsBumbleflower

func newMsBumbleflower() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Ms. Bumbleflower",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Green, color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Rabbit, types.Citizen},
			Power:      opt.Val(game.PT{Value: 1}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventSpellCast,
							Controller: game.TriggerControllerYou,
						},
					},
					CountsResolutionsThisTurn: true,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(1),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(1)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Flying,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(2),
									Player: game.ControllerReference(),
								},
								Condition: opt.Val(game.EffectCondition{
									Condition: opt.Val(game.Condition{
										SourceAbilityResolutionOrdinalThisTurn: 2,
									}),
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			Whenever you cast a spell, target opponent draws a card. Put a +1/+1 counter on target creature. It gains flying until end of turn. If this is the second time this ability has resolved this turn, you draw two cards.
		`,
		},
	}
}
