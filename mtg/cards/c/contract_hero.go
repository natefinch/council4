package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ContractHero is the card definition for Contract Hero.
//
// Type: Creature — Human Mercenary Hero
// Cost: {2}{R}
//
// Oracle text:
//
//	When this creature enters, create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")
//	Whenever this creature attacks, you may sacrifice an artifact or discard a card. If you do, this creature gets +2/+0 until end of turn.
var ContractHero = newContractHero()

func newContractHero() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Contract Hero",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Mercenary, types.Hero},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
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
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenDef(contractHeroToken),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.SacrificePermanents{
									Amount:    game.Fixed(1),
									Player:    game.ControllerReference(),
									Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
								},
								Optional:      true,
								PublishResult: game.ResultKey("disjunctive-cost-a"),
							},
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:      "disjunctive-cost-a",
									Accepted: game.TriFalse,
								}),
								Optional:      true,
								PublishResult: game.ResultKey("disjunctive-cost-b"),
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-a",
									Succeeded: game.TriTrue,
								}),
							},
							{
								Primitive: game.ModifyPT{
									Object:         game.SourcePermanentReference(),
									PowerDelta:     game.Fixed(2),
									ToughnessDelta: game.Fixed(0),
									Duration:       game.DurationUntilEndOfTurn,
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "disjunctive-cost-b",
									Succeeded: game.TriTrue,
								}),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, create a Treasure token. (It's an artifact with "{T}, Sacrifice this token: Add one mana of any color.")
			Whenever this creature attacks, you may sacrifice an artifact or discard a card. If you do, this creature gets +2/+0 until end of turn.
		`,
		},
	}
}

var contractHeroToken = newContractHeroToken()

func newContractHeroToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Treasure",
			Types:    []types.Card{types.Artifact},
			Subtypes: []types.Sub{types.Treasure},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: []cost.Additional{
						{
							Kind: cost.AdditionalTap,
						},
						{
							Kind:               cost.AdditionalSacrificeSource,
							Text:               "Sacrifice this artifact",
							Amount:             1,
							MatchPermanentType: true,
							PermanentType:      types.Artifact,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Choose{
									Choice: game.ResolutionChoice{
										Kind:   game.ResolutionChoiceMana,
										Prompt: "Choose a color",
										Colors: []mana.Color{mana.W, mana.U, mana.B, mana.R, mana.G},
									},
									PublishChoice: game.ChoiceKey("oracle-mana-color"),
								},
							},
							{
								Primitive: game.AddMana{
									Amount:     game.Fixed(1),
									ChoiceFrom: game.ChoiceKey("oracle-mana-color"),
								},
							},
						},
					}.Ability(),
				},
			},
		},
	}
}
