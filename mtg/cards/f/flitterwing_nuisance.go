package f

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// FlitterwingNuisance is the card definition for Flitterwing Nuisance.
//
// Type: Creature — Faerie Rogue
// Cost: {U}
//
// Oracle text:
//
//	Flying
//	This creature enters with a -1/-1 counter on it.
//	{2}{U}, Remove a counter from this creature: Whenever a creature you control deals combat damage to a player or planeswalker this turn, draw a card.
var FlitterwingNuisance = newFlitterwingNuisance()

func newFlitterwingNuisance() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Flitterwing Nuisance",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Rogue},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text:     "{2}{U}, Remove a counter from this creature: Whenever a creature you control deals combat damage to a player or planeswalker this turn, draw a card.",
					ManaCost: opt.Val(cost.Mana{cost.O(2), cost.U}),
					AdditionalCosts: []cost.Additional{
						{
							Kind:           cost.AdditionalRemoveCounter,
							Text:           "Remove a counter from this creature",
							Amount:         1,
							AnyCounterKind: true,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										EventPattern: opt.Val(game.TriggerPattern{
											Event:                    game.EventDamageDealt,
											Controller:               game.TriggerControllerYou,
											Subject:                  game.TriggerSubjectDamageSource,
											RequireCombatDamage:      true,
											DamageRecipient:          game.DamageRecipientPlayer | game.DamageRecipientPermanent,
											DamageRecipientSelection: game.Selection{RequiredTypes: []types.Card{types.Planeswalker}},
											DamageSourceSelection:    game.Selection{RequiredTypes: []types.Card{types.Creature}},
										}),
										Window: game.DelayedWindowThisTurn,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Draw{
														Amount: game.Fixed(1),
														Player: game.ControllerReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This creature enters with a -1/-1 counter on it.", game.CounterPlacement{Kind: counter.MinusOneMinusOne, Amount: 1}),
			},
			OracleText: `
			Flying
			This creature enters with a -1/-1 counter on it.
			{2}{U}, Remove a counter from this creature: Whenever a creature you control deals combat damage to a player or planeswalker this turn, draw a card.
		`,
		},
	}
}
