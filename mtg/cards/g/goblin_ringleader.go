package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinRingleader is the card definition for Goblin Ringleader.
//
// Type: Creature — Goblin
// Cost: {3}{R}
//
// Oracle text:
//
//	Haste (This creature can attack and {T} as soon as it comes under your control.)
//	When this creature enters, reveal the top four cards of your library. Put all Goblin cards revealed this way into your hand and the rest on the bottom of your library in any order.
var GoblinRingleader = newGoblinRingleader

func newGoblinRingleader() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Ringleader",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.RevealTopPartition{
									Player:    game.ControllerReference(),
									Amount:    game.Fixed(4),
									Selection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Goblin")}},
									Remainder: game.DigRemainderLibraryBottom,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Haste (This creature can attack and {T} as soon as it comes under your control.)
			When this creature enters, reveal the top four cards of your library. Put all Goblin cards revealed this way into your hand and the rest on the bottom of your library in any order.
		`,
		},
	}
}
