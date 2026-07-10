package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrazenCollector is the card definition for Brazen Collector.
//
// Type: Creature — Raccoon Rogue
// Cost: {1}{R}
//
// Oracle text:
//
//	First strike
//	Whenever this creature attacks, add {R}. Until end of turn, you don't lose this mana as steps and phases end.
var BrazenCollector = newBrazenCollector

func newBrazenCollector() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Brazen Collector",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Raccoon, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:                game.Fixed(1),
									ManaColor:             mana.R,
									PersistUntilEndOfTurn: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			First strike
			Whenever this creature attacks, add {R}. Until end of turn, you don't lose this mana as steps and phases end.
		`,
		},
	}
}
