package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// IngeniousInfiltrator is the card definition for Ingenious Infiltrator.
//
// Type: Creature — Vedalken Ninja
// Cost: {2}{U}{B}
//
// Oracle text:
//
//	Ninjutsu {U}{B} ({U}{B}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)
//	Whenever a Ninja you control deals combat damage to a player, draw a card.
var IngeniousInfiltrator = newIngeniousInfiltrator

func newIngeniousInfiltrator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Ingenious Infiltrator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
				cost.B,
			}),
			Colors:    []color.Color{color.Black, color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Ninja},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "Ninjutsu {U}{B}",
					ManaCost: opt.Val(cost.Mana{
						cost.U,
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalReturnUnblockedAttacker,
							Text:        "Return an unblocked attacker you control to its owner's hand",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Hand,
					Timing:         game.DuringCombat,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{},
						},
						MinModes: 1,
						MaxModes: 1,
					},
					KeywordAbilities: []game.KeywordAbility{
						game.NinjutsuKeyword{
							Cost: cost.Mana{
								cost.U,
								cost.B,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a Ninja you control deals combat damage to a player, draw a card.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							StackObjectKind:       game.StackSpell,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{SubtypesAny: []types.Sub{types.Ninja}},
							RequireCombatDamage:   true,
							CounterKind:           counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
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
			Ninjutsu {U}{B} ({U}{B}, Return an unblocked attacker you control to hand: Put this card onto the battlefield from your hand tapped and attacking.)
			Whenever a Ninja you control deals combat damage to a player, draw a card.
		`,
		},
	}
}
