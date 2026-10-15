package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// Spirit is the card definition for Spirit.
//
// Type: Token Creature — Spirit
//
// Oracle text:
//   Vigilance
//   Whenever you draw a card, put a +1/+1 counter on this creature.

// SpiritToken2c6d6effecb24b988e1fc5a5cdf95ac5 is the card definition for Spirit.
var SpiritToken2c6d6effecb24b988e1fc5a5cdf95ac5 = newSpiritToken2c6d6effecb24b988e1fc5a5cdf95ac5()

func newSpiritToken2c6d6effecb24b988e1fc5a5cdf95ac5() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:      "Spirit",
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.VigilanceStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever you draw a card, put a +1/+1 counter on this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventCardDrawn,
							Player:          game.TriggerPlayerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			OracleText: `
			Vigilance
			Whenever you draw a card, put a +1/+1 counter on this creature.
		`,
		},
	}
}
