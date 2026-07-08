package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TentativeConnection is the card definition for Tentative Connection.
//
// Type: Sorcery
// Cost: {3}{R}
//
// Oracle text:
//
//	This spell costs {3} less to cast if you control a creature with menace.
//	Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.
var TentativeConnection = newTentativeConnection

func newTentativeConnection() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Tentative Connection",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 3,
								ReductionCondition: opt.Val(game.Condition{
									ControlsMatching: opt.Val(game.SelectionCount{
										Selection: game.Selection{RequiredTypes: []types.Card{types.Creature}, Keyword: game.Menace},
									}),
								}),
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
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
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer:         game.LayerControl,
									NewController: opt.Val(game.Player1),
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
					{
						Primitive: game.Untap{
							Object: game.TargetPermanentReference(0),
						},
					},
					{
						Primitive: game.ApplyContinuous{
							Object: opt.Val(game.TargetPermanentReference(0)),
							ContinuousEffects: []game.ContinuousEffect{
								game.ContinuousEffect{
									Layer: game.LayerAbility,
									AddKeywords: []game.Keyword{
										game.Haste,
									},
								},
							},
							Duration: game.DurationUntilEndOfTurn,
						},
					},
				},
			}.Ability()),
			OracleText: `
			This spell costs {3} less to cast if you control a creature with menace.
			Gain control of target creature until end of turn. Untap that creature. It gains haste until end of turn.
		`,
		},
	}
}
