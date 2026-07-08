package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AxelrodGunnarson is the card definition for Axelrod Gunnarson.
//
// Type: Legendary Creature — Giant
// Cost: {4}{B}{B}{R}{R}
//
// Oracle text:
//
//	Trample
//	Whenever a creature dealt damage by Axelrod Gunnarson this turn dies, you gain 1 life and Axelrod Gunnarson deals 1 damage to target player or planeswalker.
var AxelrodGunnarson = newAxelrodGunnarson

func newAxelrodGunnarson() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black, color.Red),
		CardFace: game.CardFace{
			Name: "Axelrod Gunnarson",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.B,
				cost.B,
				cost.R,
				cost.R,
			}),
			Colors:     []color.Color{color.Black, color.Red},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Giant},
			Power:      opt.Val(game.PT{Value: 5}),
			Toughness:  opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.TrampleStaticBody,
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
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target player or planeswalker",
								Allow:      game.TargetAllowPermanent | game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Planeswalker}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(1),
									Recipient: game.AnyTargetDamageRecipient(0),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Trample
			Whenever a creature dealt damage by Axelrod Gunnarson this turn dies, you gain 1 life and Axelrod Gunnarson deals 1 damage to target player or planeswalker.
		`,
		},
	}
}
