package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GodPharaohSStatue is the card definition for God-Pharaoh's Statue.
//
// Type: Legendary Artifact
// Cost: {6}
//
// Oracle text:
//
//	Spells your opponents cast cost {2} more to cast.
//	At the beginning of your end step, each opponent loses 1 life.
var GodPharaohSStatue = newGodPharaohSStatue()

func newGodPharaohSStatue() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "God-Pharaoh's Statue",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedPlayer: game.PlayerOpponent,
							CostModifier: game.CostModifier{
								Kind:            game.CostModifierSpell,
								GenericIncrease: 2,
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
							Step:       game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount:      game.Fixed(1),
									PlayerGroup: game.OpponentsReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Spells your opponents cast cost {2} more to cast.
			At the beginning of your end step, each opponent loses 1 life.
		`,
		},
	}
}
