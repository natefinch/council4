package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuicksilverBehemoth is the card definition for Quicksilver Behemoth.
//
// Type: Creature — Beast
// Cost: {6}{U}
//
// Oracle text:
//
//	Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
//	When this creature attacks or blocks, return it to its owner's hand at end of combat. (Return it only if it's on the battlefield.)
var QuicksilverBehemoth = newQuicksilverBehemoth()

func newQuicksilverBehemoth() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Quicksilver Behemoth",
			ManaCost: opt.Val(cost.Mana{
				cost.O(6),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Beast},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:               game.CostModifierSpell,
								PerObjectReduction: 1,
								CountSelection:     &game.Selection{RequiredTypes: []types.Card{types.Artifact}, Controller: game.ControllerYou},
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:      game.EventAttackerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventBlockerDeclared,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtEndOfCombat,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Bounce{
														Object: game.SourceCardPermanentReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Affinity for artifacts (This spell costs {1} less to cast for each artifact you control.)
			When this creature attacks or blocks, return it to its owner's hand at end of combat. (Return it only if it's on the battlefield.)
		`,
		},
	}
}
