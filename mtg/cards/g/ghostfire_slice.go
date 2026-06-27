package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GhostfireSlice is the card definition for Ghostfire Slice.
//
// Type: Instant
// Cost: {2}{R}
//
// Oracle text:
//
//	Devoid (This card has no color.)
//	This spell costs {2} less to cast if an opponent controls a multicolored permanent.
//	Ghostfire Slice deals 4 damage to any target.
var GhostfireSlice = newGhostfireSlice()

func newGhostfireSlice() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ghostfire Slice",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Types: []types.Card{types.Instant},
			StaticAbilities: []game.StaticAbility{
				game.DevoidStaticBody,
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							AffectedSource: true,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								ReductionCondition: opt.Val(game.Condition{
									AnyOpponentControls: opt.Val(game.SelectionCount{
										Selection: game.Selection{Multicolored: true},
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
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
				},
			}.Ability()),
			OracleText: `
			Devoid (This card has no color.)
			This spell costs {2} less to cast if an opponent controls a multicolored permanent.
			Ghostfire Slice deals 4 damage to any target.
		`,
		},
	}
}
