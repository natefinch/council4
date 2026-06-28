package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CoralAtoll is the card definition for Coral Atoll.
//
// Type: Land
//
// Oracle text:
//
//	This land enters tapped.
//	When this land enters, sacrifice it unless you return an untapped Island you control to its owner's hand.
//	{T}: Add {C}{U}.
var CoralAtoll = newCoralAtoll()

func newCoralAtoll() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name:  "Coral Atoll",
			Types: []types.Card{types.Land},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					AdditionalCosts: cost.Tap,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.C,
								},
							},
							{
								Primitive: game.AddMana{
									Amount:    game.Fixed(1),
									ManaColor: mana.U,
								},
							},
						},
					}.Ability(),
				},
			},
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
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Return an untapped Island you control to its owner's hand?",
										AdditionalCosts: []cost.Additional{
											{
												Kind:            cost.AdditionalReturnToHand,
												Text:            "return an untapped Island you control to its owner's hand",
												Amount:          1,
												RequireUntapped: true,
												SubtypesAny:     cost.SubtypeSet{types.Island},
											},
										},
									},
								},
								PublishResult: game.ResultKey("sacrifice-unless-paid"),
							},
							{
								Primitive: game.Sacrifice{
									Object: game.EventPermanentReference(),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "sacrifice-unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersTappedReplacement("This land enters tapped."),
			},
			OracleText: `
			This land enters tapped.
			When this land enters, sacrifice it unless you return an untapped Island you control to its owner's hand.
			{T}: Add {C}{U}.
		`,
		},
	}
}
