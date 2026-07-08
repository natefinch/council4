package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RelicGolem is the card definition for Relic Golem.
//
// Type: Artifact Creature — Golem
// Cost: {3}
//
// Oracle text:
//
//	This creature can't attack or block unless an opponent has eight or more cards in their graveyard.
//	{2}, {T}: Target player mills two cards. (They put the top two cards of their library into their graveyard.)
var RelicGolem = newRelicGolem

func newRelicGolem() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Relic Golem",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Condition: opt.Val(game.Condition{
						Negate:     true,
						Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateAnyOpponentGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 8}},
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCantAttack,
							AffectedSource: true,
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCantBlock,
							AffectedSource: true,
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:            "{2}, {T}: Target player mills two cards. (They put the top two cards of their library into their graveyard.)",
					ManaCost:        opt.Val(cost.Mana{cost.O(2)}),
					AdditionalCosts: cost.Tap,
					ZoneOfFunction:  zone.Battlefield,
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target player",
								Allow:      game.TargetAllowPlayer,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Mill{
									Amount: game.Fixed(2),
									Player: game.TargetPlayerReference(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			This creature can't attack or block unless an opponent has eight or more cards in their graveyard.
			{2}, {T}: Target player mills two cards. (They put the top two cards of their library into their graveyard.)
		`,
		},
	}
}
