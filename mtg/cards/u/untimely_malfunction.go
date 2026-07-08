package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UntimelyMalfunction is the card definition for Untimely Malfunction.
//
// Type: Instant
// Cost: {1}{R}
//
// Oracle text:
//
//	Choose one —
//	• Destroy target artifact.
//	• Change the target of target spell or ability with a single target.
//	• One or two target creatures can't block this turn.
var UntimelyMalfunction = newUntimelyMalfunction

func newUntimelyMalfunction() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Untimely Malfunction",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Destroy target artifact.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Change the target of target spell or ability with a single target.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target spell or ability",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackSpell, game.StackActivatedAbility, game.StackTriggeredAbility},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ChooseNewTargets{
									Object: game.TargetStackObjectReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "One or two target creatures can't block this turn.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 2,
								Constraint: "one or two target creatures",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
							{
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(1)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Destroy target artifact.
			• Change the target of target spell or ability with a single target.
			• One or two target creatures can't block this turn.
		`,
		},
	}
}
