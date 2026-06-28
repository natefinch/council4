package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BriberSPurse is the card definition for Briber's Purse.
//
// Type: Artifact
// Cost: {X}
//
// Oracle text:
//
//	This artifact enters with X gem counters on it.
//	{1}, {T}, Remove a gem counter from this artifact: Target creature can't attack or block this turn.
var BriberSPurse = newBriberSPurse()

func newBriberSPurse() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Briber's Purse",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{1}, {T}, Remove a gem counter from this artifact: Target creature can't attack or block this turn.",
					ManaCost: opt.Val(cost.Mana{cost.O(1)}),
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:        cost.AdditionalRemoveCounter,
							Text:        "Remove a gem counter from this artifact",
							Amount:      1,
							CounterKind: counter.Gem,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
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
								Primitive: game.ApplyRule{
									Object: opt.Val(game.TargetPermanentReference(0)),
									RuleEffects: []game.RuleEffect{
										game.RuleEffect{
											Kind: game.RuleEffectCantAttack,
										},
										game.RuleEffect{
											Kind: game.RuleEffectCantBlock,
										},
									},
									Duration: game.DurationThisTurn,
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with X gem counters on it.", game.CounterPlacement{Kind: counter.Gem, AmountFromX: true}),
			},
			OracleText: `
			This artifact enters with X gem counters on it.
			{1}, {T}, Remove a gem counter from this artifact: Target creature can't attack or block this turn.
		`,
		},
	}
}
