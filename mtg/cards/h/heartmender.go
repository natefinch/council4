package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Heartmender is the card definition for Heartmender.
//
// Type: Creature — Elemental
// Cost: {2}{G/W}{G/W}
//
// Oracle text:
//
//	At the beginning of your upkeep, remove a -1/-1 counter from each creature you control.
//	Persist (When this creature dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)
var Heartmender = newHeartmender()

func newHeartmender() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Heartmender",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.HybridMana(mana.G, mana.W),
				cost.HybridMana(mana.G, mana.W),
			}),
			Colors:    []color.Color{color.Green, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elemental},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:      game.EventBeginningOfStep,
							Controller: game.TriggerControllerYou,
							Step:       game.StepUpkeep,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RemoveCounter{
									Amount:      game.Fixed(1),
									Group:       game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
									CounterKind: counter.MinusOneMinusOne,
								},
							},
						},
					}.Ability(),
				},
				game.PersistTriggeredBody,
			},
			OracleText: `
			At the beginning of your upkeep, remove a -1/-1 counter from each creature you control.
			Persist (When this creature dies, if it had no -1/-1 counters on it, return it to the battlefield under its owner's control with a -1/-1 counter on it.)
		`,
		},
	}
}
