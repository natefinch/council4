package j

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// JeditOjanenMercenary is the card definition for Jedit Ojanen, Mercenary.
//
// Type: Legendary Creature — Cat Mercenary
// Cost: {1}{W}{U}
//
// Oracle text:
//
//	Whenever Jedit Ojanen or another legendary creature you control enters, you may pay {G}. If you do, create a 2/2 green Cat Warrior creature token with forestwalk. (It can't be blocked as long as defending player controls a Forest.)
var JeditOjanenMercenary = newJeditOjanenMercenary

func newJeditOjanenMercenary() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Jedit Ojanen, Mercenary",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cat, types.Mercenary},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever Jedit Ojanen or another legendary creature you control enters, you may pay {G}. If you do, create a 2/2 green Cat Warrior creature token with forestwalk. (It can't be blocked as long as defending player controls a Forest.)",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                  game.EventPermanentEnteredBattlefield,
							Controller:             game.TriggerControllerYou,
							SubjectSelection:       game.Selection{RequiredTypes: []types.Card{types.Creature}, Supertypes: []types.Super{types.Legendary}},
							SubjectSelectionOrSelf: true,
							StackObjectKind:        game.StackSpell,
							CounterKind:            counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Pay{
											Payment: game.ResolutionPayment{
												Prompt: "Pay {G}?",
												ManaCost: opt.Val(cost.Mana{
													cost.G,
												}),
											},
										},
										PublishResult: game.ResultKey("controller-paid"),
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(1),
											Source: game.TokenDef(jeditOjanenMercenaryToken),
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
			Whenever Jedit Ojanen or another legendary creature you control enters, you may pay {G}. If you do, create a 2/2 green Cat Warrior creature token with forestwalk. (It can't be blocked as long as defending player controls a Forest.)
		`,
		},
	}
}

var jeditOjanenMercenaryToken = newJeditOjanenMercenaryToken()

func newJeditOjanenMercenaryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Cat Warrior",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Cat, types.Warrior},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.ForestwalkStaticBody,
			},
		},
	}
}
