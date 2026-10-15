package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// LeelaSevateemWarrior is the card definition for Leela, Sevateem Warrior.
//
// Type: Legendary Creature — Human Warrior
// Cost: {3}{G}
//
// Oracle text:
//
//	Whenever an opponent draws a card except the first one they draw in each of their draw steps, put a +1/+1 counter on Leela.
//	Doctor's companion (You can have two commanders if the other is the Doctor.)
var LeelaSevateemWarrior = newLeelaSevateemWarrior

func newLeelaSevateemWarrior() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Leela, Sevateem Warrior",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Human, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.CompanionStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever an opponent draws a card except the first one they draw in each of their draw steps, put a +1/+1 counter on Leela.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                      game.EventCardDrawn,
							Player:                     game.TriggerPlayerOpponent,
							StackObjectKind:            game.StackSpell,
							ExcludeFirstDrawInDrawStep: true,
							CounterKind:                counter.PlusOnePlusOne,
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
			Whenever an opponent draws a card except the first one they draw in each of their draw steps, put a +1/+1 counter on Leela.
			Doctor's companion (You can have two commanders if the other is the Doctor.)
		`,
		},
	}
}
