package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JackedRabbit is the card definition for Jacked Rabbit.
//
// Type: Creature — Rabbit Warrior
// Cost: {X}{1}{W}
//
// Oracle text:
//
//	Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)
//	Whenever this creature attacks, create a number of 1/1 white Rabbit creature tokens equal to this creature's power.
var JackedRabbit = newJackedRabbit

func newJackedRabbit() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White),
		CardFace: game.CardFace{
			Name: "Jacked Rabbit",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.O(1),
				cost.W,
			}),
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit, types.Warrior},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.RavenousDrawTriggeredAbility(),
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, create a number of 1/1 white Rabbit creature tokens equal to this creature's power.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountObjectPower,
												Multiplier: 1,
												Object:     game.SourcePermanentReference(),
											}),
											Source: game.TokenDef(jackedRabbitToken),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.RavenousEntersWithCountersReplacement(),
			},
			OracleText: `
			Ravenous (This creature enters with X +1/+1 counters on it. If X is 5 or more, draw a card when it enters.)
			Whenever this creature attacks, create a number of 1/1 white Rabbit creature tokens equal to this creature's power.
		`,
		},
	}
}

var jackedRabbitToken = newJackedRabbitToken()

func newJackedRabbitToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Rabbit",
			Colors:    []color.Color{color.White},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Rabbit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
