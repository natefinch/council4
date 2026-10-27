// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ItThatBetrays is the card definition for It That Betrays.
//
// Type: Creature — Eldrazi
// Cost: {12}
//
// Oracle text:
//
//	Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
//	Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control.
var ItThatBetrays = newItThatBetrays

func newItThatBetrays() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "It That Betrays",
			ManaCost: opt.Val(cost.Mana{
				cost.O(12),
			}),
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi},
			Power:     opt.Val(game.PT{Value: 11}),
			Toughness: opt.Val(game.PT{Value: 11}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, defending player sacrifices two permanents of their choice.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:           game.EventAttackerDeclared,
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
										Primitive: game.SacrificePermanents{
											Player: game.DefendingPlayerReference(),
											Amount: game.Fixed(2),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.TriggeredAbility{
					Text: "Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentSacrificed,
							Player:           game.TriggerPlayerOpponent,
							SubjectSelection: game.Selection{NonToken: true},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.PutOnBattlefield{
											Source:    game.CardBattlefieldSource(game.CardReference{Kind: game.CardReferenceEvent}),
											Recipient: opt.Val(game.ControllerReference()),
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
			Annihilator 2 (Whenever this creature attacks, defending player sacrifices two permanents of their choice.)
			Whenever an opponent sacrifices a nontoken permanent, put that card onto the battlefield under your control.
		`,
		},
	}
}
