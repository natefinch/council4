package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CabalRitual is the card definition for Cabal Ritual.
//
// Type: Instant
// Cost: {1}{B}
//
// Oracle text:
//
//	Add {B}{B}{B}.
//	Threshold — Add {B}{B}{B}{B}{B} instead if there are seven or more cards in your graveyard.
var CabalRitual = newCabalRitual

func newCabalRitual() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Cabal Ritual",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:     true,
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardCardCount, Op: compare.GreaterOrEqual, Value: 7}},
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate: "condition-0",
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate: "condition-0",
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.AddMana{
							Amount:    game.Fixed(1),
							ManaColor: mana.B,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Add {B}{B}{B}.
			Threshold — Add {B}{B}{B}{B}{B} instead if there are seven or more cards in your graveyard.
		`,
		},
	}
}
