package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UnholyHeat is the card definition for Unholy Heat.
//
// Type: Instant
// Cost: {R}
//
// Oracle text:
//
//	Unholy Heat deals 2 damage to target creature or planeswalker.
//	Delirium — Unholy Heat deals 6 damage instead if there are four or more card types among cards in your graveyard.
var UnholyHeat = newUnholyHeat

func newUnholyHeat() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Unholy Heat",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "target creature or planeswalker",
						Allow:      game.TargetAllowPermanent,
						Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(2),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:     true,
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardTypeCount, Op: compare.GreaterOrEqual, Value: 4}},
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(6),
							Recipient: game.AnyTargetDamageRecipient(0),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Unholy Heat deals 2 damage to target creature or planeswalker.
			Delirium — Unholy Heat deals 6 damage instead if there are four or more card types among cards in your graveyard.
		`,
		},
	}
}
