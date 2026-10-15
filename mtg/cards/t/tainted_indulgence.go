package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/compare"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TaintedIndulgence is the card definition for Tainted Indulgence.
//
// Type: Instant
// Cost: {U}{B}
//
// Oracle text:
//
//	Draw two cards. Then discard a card unless there are five or more mana values among cards in your graveyard.
var TaintedIndulgence = newTaintedIndulgence

func newTaintedIndulgence() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Tainted Indulgence",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Draw{
							Amount: game.Fixed(2),
							Player: game.ControllerReference(),
						},
					},
					{
						Primitive: game.Discard{
							Amount: game.Fixed(1),
							Player: game.ControllerReference(),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:     true,
								Aggregates: []game.AggregateComparison{{Aggregate: game.AggregateControllerGraveyardManaValueCount, Op: compare.GreaterOrEqual, Value: 5}},
							}),
						}),
					},
				},
			}.Ability()),
			OracleText: `
			Draw two cards. Then discard a card unless there are five or more mana values among cards in your graveyard.
		`,
		},
	}
}
