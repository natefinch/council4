package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RadiantSDragoons is the card definition for Radiant's Dragoons.
//
// Type: Creature — Human Soldier
// Cost: {3}{W}
//
// Oracle text:
//
//	Echo {3}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters, you gain 5 life.
var RadiantSDragoons = newRadiantSDragoons

func newRadiantSDragoons() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Radiant's Dragoons",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Soldier},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.O(3), cost.W}),
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
								Primitive: game.GainLife{
									Amount: game.Fixed(5),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {3}{W} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters, you gain 5 life.
		`,
		},
	}
}
