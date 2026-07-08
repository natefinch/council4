package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HallowedHealer is the card definition for Hallowed Healer.
//
// Type: Creature — Human Cleric
// Cost: {2}{W}
//
// Oracle text:
//
//	{T}: Prevent the next 2 damage that would be dealt to any target this turn.
//	Threshold — {T}: Prevent the next 4 damage that would be dealt to any target this turn. Activate only if there are seven or more cards in your graveyard.
var HallowedHealer = newHallowedHealer

func newHallowedHealer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Hallowed Healer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Cleric},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{T}: Prevent the next 2 damage that would be dealt to any target this turn.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
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
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(2),
								},
							},
						},
					}.Ability(),
				},
				game.ActivatedAbility{
					Text:            "Threshold — {T}: Prevent the next 4 damage that would be dealt to any target this turn. Activate only if there are seven or more cards in your graveyard.",
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					ActivationCondition: opt.Val(game.Condition{
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
					}),
					Content: game.Mode{
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
								Primitive: game.PreventDamage{
									AnyTarget: game.AnyTargetDamageRecipient(0),
									Amount:    game.Fixed(4),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			{T}: Prevent the next 2 damage that would be dealt to any target this turn.
			Threshold — {T}: Prevent the next 4 damage that would be dealt to any target this turn. Activate only if there are seven or more cards in your graveyard.
		`,
		},
	}
}
