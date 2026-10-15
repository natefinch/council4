package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CarefreeSwinemaster is the card definition for Carefree Swinemaster.
//
// Type: Creature — Gnome Ranger
// Cost: {2}{G}
//
// Oracle text:
//
//	Whenever this creature attacks, you may pay {1}{G}. If you do, create a 2/2 green Boar creature token that's tapped and attacking.
var CarefreeSwinemaster = newCarefreeSwinemaster

func newCarefreeSwinemaster() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Carefree Swinemaster",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Gnome, types.Ranger},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 4}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, you may pay {1}{G}. If you do, create a 2/2 green Boar creature token that's tapped and attacking.",
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
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {1}{G}?",
												ManaCost: opt.Val(cost.Mana{
													cost.O(1),
													cost.G,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount:         game.Fixed(1),
											Source:         game.TokenDef(carefreeSwinemasterToken),
											EntryTapped:    true,
											EntryAttacking: true,
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:       game.ResultKey("controller-paid"),
											Succeeded: game.TriTrue,
										}),
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
			Whenever this creature attacks, you may pay {1}{G}. If you do, create a 2/2 green Boar creature token that's tapped and attacking.
		`,
		},
	}
}

var carefreeSwinemasterToken = newCarefreeSwinemasterToken()

func newCarefreeSwinemasterToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Boar",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Boar},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
		},
	}
}
