package i

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ImpetuousProtege is the card definition for Impetuous Protege.
//
// Type: Creature — Human Warrior
// Cost: {2}{R}
//
// Oracle text:
//
//	Partner with Proud Mentor (When this creature enters, target player may put Proud Mentor into their hand from their library, then shuffle.)
//	Whenever this creature attacks, it gets +X/+0 until end of turn, where X is the greatest power among tapped creatures your opponents control.
var ImpetuousProtege = newImpetuousProtege

func newImpetuousProtege() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Impetuous Protege",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Warrior},
			Power:     opt.Val(game.PT{Value: 0}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.PartnerWithStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever this creature attacks, it gets +X/+0 until end of turn, where X is the greatest power among tapped creatures your opponents control.",
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
										Primitive: game.ModifyPT{
											Object: game.EventPermanentReference(),
											PowerDelta: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountGreatestPowerInGroup,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent, Tapped: game.TriTrue}),
											}),
											Duration: game.DurationUntilEndOfTurn,
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
				game.PartnerWithTriggeredAbility("Proud Mentor"),
			},
			OracleText: `
			Partner with Proud Mentor (When this creature enters, target player may put Proud Mentor into their hand from their library, then shuffle.)
			Whenever this creature attacks, it gets +X/+0 until end of turn, where X is the greatest power among tapped creatures your opponents control.
		`,
		},
	}
}
