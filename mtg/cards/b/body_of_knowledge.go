package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BodyOfKnowledge is the card definition for Body of Knowledge.
//
// Type: Creature — Avatar
// Cost: {3}{U}{U}
//
// Oracle text:
//
//	Body of Knowledge's power and toughness are each equal to the number of cards in your hand.
//	You have no maximum hand size.
//	Whenever this creature is dealt damage, draw that many cards.
var BodyOfKnowledge = newBodyOfKnowledge

func newBodyOfKnowledge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Body of Knowledge",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
				cost.U,
			}),
			Colors:           []color.Color{color.Blue},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Avatar},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature is dealt damage, draw that many cards.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventDamageDealt,
							Source:          game.TriggerSourceSelf,
							Subject:         game.TriggerSubjectPermanent,
							StackObjectKind: game.StackSpell,
							DamageRecipient: game.DamageRecipientPermanent,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountEventDamage,
												Multiplier: 1,
											}),
											Player: game.ControllerReference(),
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
			Body of Knowledge's power and toughness are each equal to the number of cards in your hand.
			You have no maximum hand size.
			Whenever this creature is dealt damage, draw that many cards.
		`,
		},
	}
}
