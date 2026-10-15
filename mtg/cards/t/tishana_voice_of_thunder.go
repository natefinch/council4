package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// TishanaVoiceOfThunder is the card definition for Tishana, Voice of Thunder.
//
// Type: Legendary Creature — Merfolk Shaman
// Cost: {5}{G}{U}
//
// Oracle text:
//
//	Tishana's power and toughness are each equal to the number of cards in your hand.
//	You have no maximum hand size.
//	When Tishana enters, draw a card for each creature you control.
var TishanaVoiceOfThunder = newTishanaVoiceOfThunder

func newTishanaVoiceOfThunder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Green),
		CardFace: game.CardFace{
			Name: "Tishana, Voice of Thunder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
				cost.U,
			}),
			Colors:           []color.Color{color.Green, color.Blue},
			Supertypes:       []types.Super{types.Legendary},
			Types:            []types.Card{types.Creature},
			Subtypes:         []types.Sub{types.Merfolk, types.Shaman},
			Power:            opt.Val(game.PT{IsStar: true}),
			Toughness:        opt.Val(game.PT{IsStar: true}),
			DynamicPower:     opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			DynamicToughness: opt.Val(game.DynamicValue{Kind: game.DynamicValueControllerHandSize}),
			StaticAbilities: []game.StaticAbility{
				game.NoMaximumHandSizeStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When Tishana enters, draw a card for each creature you control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
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
										Primitive: game.Draw{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
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
			Tishana's power and toughness are each equal to the number of cards in your hand.
			You have no maximum hand size.
			When Tishana enters, draw a card for each creature you control.
		`,
		},
	}
}
