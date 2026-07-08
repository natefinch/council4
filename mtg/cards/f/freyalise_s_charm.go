package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FreyaliseSCharm is the card definition for Freyalise's Charm.
//
// Type: Enchantment
// Cost: {G}{G}
//
// Oracle text:
//
//	Whenever an opponent casts a black spell, you may pay {G}{G}. If you do, you draw a card.
//	{G}{G}: Return this enchantment to its owner's hand.
var FreyaliseSCharm = newFreyaliseSCharm

func newFreyaliseSCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Freyalise's Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:           "{G}{G}: Return this enchantment to its owner's hand.",
					ManaCost:       opt.Val(cost.Mana{cost.G, cost.G}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Bounce{
									Object: game.SourcePermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerOpponent,
							CardSelection: game.Selection{ColorsAny: []color.Color{color.Black}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {G}{G}?",
										ManaCost: opt.Val(cost.Mana{
											cost.G,
											cost.G,
										}),
									},
								},
								PublishResult: game.ResultKey("controller-paid"),
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
			OracleText: `
			Whenever an opponent casts a black spell, you may pay {G}{G}. If you do, you draw a card.
			{G}{G}: Return this enchantment to its owner's hand.
		`,
		},
	}
}
