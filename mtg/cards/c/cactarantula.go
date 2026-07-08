package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Cactarantula is the card definition for Cactarantula.
//
// Type: Creature — Plant Spider
// Cost: {4}{G}{G}
//
// Oracle text:
//
//	This spell costs {1} less to cast if you control a Desert.
//	Reach
//	Whenever this creature becomes the target of a spell or ability an opponent controls, you may draw a card.
var Cactarantula = newCactarantula

func newCactarantula() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Cactarantula",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Plant, types.Spider},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								ReductionCondition: opt.Val(game.Condition{
									ControlsMatching: opt.Val(game.SelectionCount{
										Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Desert")}},
									}),
								}),
							},
						},
					},
				},
				game.ReachStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventObjectBecameTarget,
							Source:          game.TriggerSourceSelf,
							CauseController: game.TriggerControllerOpponent,
						},
					},
					Optional: true,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This spell costs {1} less to cast if you control a Desert.
			Reach
			Whenever this creature becomes the target of a spell or ability an opponent controls, you may draw a card.
		`,
		},
	}
}
