// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTElspethConquersDeath is the card definition for RT Elspeth Conquers Death.
//
// Type: Enchantment — Saga
// Cost: {3}{W}{W}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Exile target permanent an opponent controls with mana value 3 or greater.
//	II — Noncreature spells your opponents cast cost {2} more to cast until your next turn.
//	III — Return target creature or planeswalker card from your graveyard to the battlefield. Put a +1/+1 counter or a loyalty counter on it.
var RTElspethConquersDeath = newRTElspethConquersDeath

func newRTElspethConquersDeath() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "RT Elspeth Conquers Death",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
				cost.W,
			}),
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Exile target permanent an opponent controls with mana value 3 or greater.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target permanent an opponent controls with mana value 3 or greater",
										Allow:      game.TargetAllowPermanent,
										Selection: opt.Val(game.Selection{Controller: game.ControllerOpponent, ManaValue: opt.Val(compare.Int{
											Op:    compare.GreaterOrEqual,
											Value: 3,
										})}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MovePermanent{
											Object:      game.TargetPermanentReference(0),
											Destination: zone.Exile,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "II — Noncreature spells your opponents cast cost {2} more to cast until your next turn.",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:           game.RuleEffectCostModifier,
													Duration:       game.DurationPermanent,
													AffectedPlayer: game.PlayerOpponent,
													CostModifier: game.CostModifier{
														Kind:            game.CostModifierSpell,
														GenericIncrease: 2,
														CardSelection:   game.Selection{ExcludedTypes: []types.Card{types.Creature}},
													},
												},
											},
											Duration: game.DurationUntilYourNextTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.ChapterAbility{
					Text:     "III — Return target creature or planeswalker card from your graveyard to the battlefield. Put a +1/+1 counter or a loyalty counter on it.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature or planeswalker card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:        game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
											PublishLinked: game.LinkedKey("sequence-effect-0-product"),
										},
										LocalProducts: game.LocalProducts{
											Links: []game.LinkedKey{game.LinkedKey("sequence-effect-0-product")},
										},
									},
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.LinkedObjectReference("sequence-effect-0-product"),
											CounterKind: counter.PlusOnePlusOne,
											KindChoices: []counter.Kind{counter.PlusOnePlusOne, counter.Loyalty},
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
			I — Exile target permanent an opponent controls with mana value 3 or greater.
			II — Noncreature spells your opponents cast cost {2} more to cast until your next turn.
			III — Return target creature or planeswalker card from your graveyard to the battlefield. Put a +1/+1 counter or a loyalty counter on it.
		`,
		},
	}
}
