package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LightningSerpent is the card definition for Lightning Serpent.
//
// Type: Creature — Elemental Serpent
// Cost: {X}{R}
//
// Oracle text:
//
//	Trample, haste
//	This creature enters with X +1/+0 counters on it.
//	At the beginning of the end step, sacrifice this creature.
var LightningSerpent = newLightningSerpent()

func newLightningSerpent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Lightning Serpent",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental, types.Serpent},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event: game.EventBeginningOfStep,
							Step:  game.StepEnd,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Sacrifice{
									Object: game.SourceCardPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with X +1/+0 counters on it.", game.CounterPlacement{Kind: counter.PlusOnePlusZero, AmountFromX: true}),
			},
			OracleText: `
			Trample, haste
			This creature enters with X +1/+0 counters on it.
			At the beginning of the end step, sacrifice this creature.
		`,
		},
	}
}
