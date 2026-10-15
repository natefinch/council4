package w

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// WhispersOfEmrakul is the card definition for Whispers of Emrakul.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Target opponent discards a card at random.
//	Delirium — If there are four or more card types among cards in your graveyard, that player discards two cards at random instead.
var WhispersOfEmrakul = newWhispersOfEmrakul

func newWhispersOfEmrakul() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Whispers of Emrakul",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.Mode{
				Targets: []game.TargetSpec{
					game.TargetSpec{
						MinTargets: 1,
						MaxTargets: 1,
						Constraint: "Target opponent",
						Allow:      game.TargetAllowPlayer,
						Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
					},
				},
				Sequence: []game.Instruction{
					{
						Primitive: game.Discard{
							Amount:   game.Fixed(1),
							Player:   game.TargetPlayerReference(0),
							AtRandom: true,
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
						Primitive: game.Discard{
							Amount:   game.Fixed(2),
							Player:   game.TargetPlayerReference(0),
							AtRandom: true,
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Target opponent discards a card at random.
			Delirium — If there are four or more card types among cards in your graveyard, that player discards two cards at random instead.
		`,
		},
	}
}
