package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GarzaZolPlagueQueen is the card definition for Garza Zol, Plague Queen.
//
// Type: Legendary Creature — Vampire Noble
// Cost: {4}{U}{B}{R}
//
// Oracle text:
//
//	Flying, haste
//	Whenever a creature dealt damage by Garza Zol this turn dies, put a +1/+1 counter on Garza Zol.
//	Whenever Garza Zol deals combat damage to a player, you may draw a card.
var GarzaZolPlagueQueen = newGarzaZolPlagueQueen

func newGarzaZolPlagueQueen() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Garza Zol, Plague Queen",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
				cost.B,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red, color.Blue},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Vampire, types.Noble},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.HasteStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                game.EventPermanentDied,
							DyingDamagedBySource: true,
							SubjectSelection:     game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.SourcePermanentReference(),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:               game.EventDamageDealt,
							Source:              game.TriggerSourceSelf,
							Subject:             game.TriggerSubjectDamageSource,
							RequireCombatDamage: true,
							DamageRecipient:     game.DamageRecipientPlayer,
						},
					},
					Optional: true,
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
			OracleText: `
			Flying, haste
			Whenever a creature dealt damage by Garza Zol this turn dies, put a +1/+1 counter on Garza Zol.
			Whenever Garza Zol deals combat damage to a player, you may draw a card.
		`,
		},
	}
}
