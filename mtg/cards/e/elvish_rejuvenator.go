package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ElvishRejuvenator is the card definition for Elvish Rejuvenator.
//
// Type: Creature — Elf Druid
// Cost: {2}{G}
//
// Oracle text:
//
//	When this creature enters, look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.
var ElvishRejuvenator = newElvishRejuvenator

func newElvishRejuvenator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Elvish Rejuvenator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Druid},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
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
								Primitive: game.Dig{
									Player:       game.ControllerReference(),
									Look:         game.Fixed(5),
									Take:         game.Fixed(1),
									Remainder:    game.DigRemainderLibraryBottom,
									Filter:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
									TakeUpTo:     true,
									Destination:  zone.Battlefield,
									EntersTapped: true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
