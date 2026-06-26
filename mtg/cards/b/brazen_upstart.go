package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BrazenUpstart is the card definition for Brazen Upstart.
//
// Type: Creature — Elf Shaman
// Cost: {R}{G}{W}
//
// Oracle text:
//
//	Vigilance
//	When this creature dies, look at the top five cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
var BrazenUpstart = newBrazenUpstart()

func newBrazenUpstart() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Brazen Upstart",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.G,
				cost.W,
			}),
			Colors:    []color.Color{color.Green, color.Red, color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Elf, types.Shaman},
			Power:     opt.Val(game.PT{Value: 4}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Dig{
									Player:    game.ControllerReference(),
									Look:      game.Fixed(5),
									Take:      game.Fixed(1),
									Remainder: game.DigRemainderLibraryBottom,
									Filter:    opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
									TakeUpTo:  true,
									Reveal:    true,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Vigilance
			When this creature dies, look at the top five cards of your library. You may reveal a creature card from among them and put it into your hand. Put the rest on the bottom of your library in a random order.
		`,
		},
	}
}
