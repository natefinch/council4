package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EndlessRanksOfHYDRA is the card definition for Endless Ranks of HYDRA.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	For each opponent, you create a 2/1 black Villain creature token with menace.
//	Whenever your commander enters or attacks, you may pay {1}{B}. If you do, return this card from your graveyard to your hand.
var EndlessRanksOfHYDRA = newEndlessRanksOfHYDRA

func newEndlessRanksOfHYDRA() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Endless Ranks of HYDRA",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							UnionEvent:       game.EventAttackerDeclared,
							SubjectSelection: game.Selection{MatchCommander: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {1}{B}?",
										ManaCost: opt.Val(cost.Mana{
											cost.O(1),
											cost.B,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceSource},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "controller-paid",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.CreateToken{
							Amount: game.Dynamic(game.DynamicAmount{
								Kind:       game.DynamicAmountOpponentCount,
								Multiplier: 1,
							}),
							Source: game.TokenDef(endlessRanksOfHYDRAToken),
						},
					},
				},
			}.Ability()),
			OracleText: `
			For each opponent, you create a 2/1 black Villain creature token with menace.
			Whenever your commander enters or attacks, you may pay {1}{B}. If you do, return this card from your graveyard to your hand.
		`,
		},
	}
}

var endlessRanksOfHYDRAToken = newEndlessRanksOfHYDRAToken()

func newEndlessRanksOfHYDRAToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Villain",
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Villain},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.MenaceStaticBody,
			},
		},
	}
}
