package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// FoundryHelix is the card definition for Foundry Helix.
//
// Type: Instant
// Cost: {1}{R}{W}
//
// Oracle text:
//
//	As an additional cost to cast this spell, sacrifice a permanent.
//	Foundry Helix deals 4 damage to any target. If the sacrificed permanent was an artifact, you gain 4 life.
var FoundryHelix = newFoundryHelix

func newFoundryHelix() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Foundry Helix",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.W,
			}),
			Colors: []color.Color{color.Red, color.White},
			Types:  []types.Card{types.Instant},
			AdditionalCosts: []cost.Additional{
				{
					Kind:       cost.AdditionalSacrifice,
					Text:       "sacrifice a permanent",
					Amount:     1,
					SubjectKey: "paid-cost-1-1",
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "any target",
						Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker, types.Battle}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
					},
					{
						Primitive: game.GainLife{
							Amount: game.Fixed(4),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Object:        opt.Val(game.PaidCostReference("paid-cost-1-1", game.PaidCostSacrifice, "")),
								ObjectMatches: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Artifact}}),
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			As an additional cost to cast this spell, sacrifice a permanent.
			Foundry Helix deals 4 damage to any target. If the sacrificed permanent was an artifact, you gain 4 life.
		`,
		},
	}
}
