package l

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// LordWindgrace is the card definition for Lord Windgrace.
//
// Type: Legendary Planeswalker — Windgrace
// Cost: {2}{B}{R}{G}
//
// Oracle text:
//
//	+2: Discard a card, then draw a card. If a land card is discarded this way, draw an additional card.
//	−3: Return up to two target land cards from your graveyard to the battlefield.
//	−11: Destroy up to six target nonland permanents, then create six 2/2 green Cat Warrior creature tokens with forestwalk.
//	Lord Windgrace can be your commander.
var LordWindgrace = newLordWindgrace

func newLordWindgrace() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red, color.Green),
		CardFace: game.CardFace{
			Name: "Lord Windgrace",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
				cost.R,
				cost.G,
			}),
			Colors:         []color.Color{color.Black, color.Green, color.Red},
			CanBeCommander: true,
			Supertypes:     []types.Super{types.Legendary},
			Types:          []types.Card{types.Planeswalker},
			Subtypes:       []types.Sub{types.Windgrace},
			Loyalty:        opt.Val(5),
			LoyaltyAbilities: []game.LoyaltyAbility{
				game.LoyaltyAbility{
					Text:        "+2: Discard a card, then draw a card. If a land card is discarded this way, draw an additional card.",
					LoyaltyCost: 2,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										PublishResult: game.ResultKey("if-you-do"),
										LocalProducts: game.LocalProducts{
											Results: []game.ResultKey{game.ResultKey("if-you-do")},
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										ResultGate: opt.Val(game.InstructionResultGate{
											Key:             game.ResultKey("if-you-do"),
											Succeeded:       game.TriTrue,
											ObjectSelection: opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
											CardOnly:        true,
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−3: Return up to two target land cards from your graveyard to the battlefield.",
					LoyaltyCost: -3,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 2,
										Constraint: "up to two target land cards from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget}),
										},
									},
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source: game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1}),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.LoyaltyAbility{
					Text:        "−11: Destroy up to six target nonland permanents, then create six 2/2 green Cat Warrior creature tokens with forestwalk.",
					LoyaltyCost: -11,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets: 6,
										Constraint: "up to six target nonland permanents",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(1),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(2),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(3),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(4),
										},
									},
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(5),
										},
									},
									game.Instruction{
										Primitive: game.CreateToken{
											Amount: game.Fixed(6),
											Source: game.TokenDef(lordWindgraceToken),
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
			+2: Discard a card, then draw a card. If a land card is discarded this way, draw an additional card.
			−3: Return up to two target land cards from your graveyard to the battlefield.
			−11: Destroy up to six target nonland permanents, then create six 2/2 green Cat Warrior creature tokens with forestwalk.
			Lord Windgrace can be your commander.
		`,
		},
	}
}

var lordWindgraceToken = newLordWindgraceToken()

func newLordWindgraceToken() *game.CardDef {
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
