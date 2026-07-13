package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MultaniSAcolyte is the card definition for Multani's Acolyte.
//
// Type: Creature — Elf
// Cost: {G}{G}
//
// Oracle text:
//
//	Echo {G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	When this creature enters, draw a card.
var MultaniSAcolyte = newMultaniSAcolyte

func newMultaniSAcolyte() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Multani's Acolyte",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.EchoTriggeredAbility(cost.Mana{cost.G, cost.G}),
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
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Echo {G}{G} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
			When this creature enters, draw a card.
		`,
		},
	}
}
