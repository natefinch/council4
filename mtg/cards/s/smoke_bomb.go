package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SmokeBomb is the card definition for Smoke Bomb.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Flash
//	All creatures have shroud. (They can't be the targets of spells or abilities.)
//	At the beginning of your upkeep, sacrifice this artifact. When you do, target creature you control can't be blocked this turn.
var SmokeBomb = newSmokeBomb()

func newSmokeBomb() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Smoke Bomb",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.FlashStaticBody,
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
							AddKeywords: []game.Keyword{
								game.Shroud,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.SourceCardPermanentReference(),
								},
							},
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBeBlocked,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flash
			All creatures have shroud. (They can't be the targets of spells or abilities.)
			At the beginning of your upkeep, sacrifice this artifact. When you do, target creature you control can't be blocked this turn.
		`,
		},
	}
}
