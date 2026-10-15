package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TheAkroanWar is the card definition for The Akroan War.
//
// Type: Enchantment — Saga
// Cost: {3}{R}
//
// Oracle text:
//
//	(As this Saga enters and after your draw step, add a lore counter. Sacrifice after III.)
//	I — Gain control of target creature for as long as this Saga remains on the battlefield.
//	II — Until your next turn, creatures your opponents control attack each combat if able.
//	III — Each tapped creature deals damage to itself equal to its power.
var TheAkroanWar = newTheAkroanWar

func newTheAkroanWar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "The Akroan War",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:   []color.Color{color.Red},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Saga},
			ChapterAbilities: []game.ChapterAbility{
				game.ChapterAbility{
					Text:     "I — Gain control of target creature for as long as this Saga remains on the battlefield.",
					Chapters: []int{1},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target creature",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyContinuous{
											Object: opt.Val(game.TargetPermanentReference(0)),
											ContinuousEffects: []game.ContinuousEffect{
												game.ContinuousEffect{
													Duration:      game.DurationPermanent,
													Layer:         game.LayerControl,
													NewController: opt.Val(game.Player1),
												},
											},
											Duration: game.DurationForAsLongAsSourceOnBattlefield,
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
					Text:     "II — Until your next turn, creatures your opponents control attack each combat if able.",
					Chapters: []int{2},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ApplyRule{
											RuleEffects: []game.RuleEffect{
												game.RuleEffect{
													Kind:               game.RuleEffectMustAttack,
													Duration:           game.DurationPermanent,
													AffectedController: game.ControllerOpponent,
													PermanentTypes:     []types.Card{types.Creature},
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
					Text:     "III — Each tapped creature deals damage to itself equal to its power.",
					Chapters: []int{3},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.GroupSelfPowerDamage{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Tapped: game.TriTrue}),
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
			I — Gain control of target creature for as long as this Saga remains on the battlefield.
			II — Until your next turn, creatures your opponents control attack each combat if able.
			III — Each tapped creature deals damage to itself equal to its power.
		`,
		},
	}
}
