package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ColossusOfTheBloodAge is the card definition for Colossus of the Blood Age.
//
// Type: Artifact Creature — Construct
// Cost: {4}{R}{W}
//
// Oracle text:
//
//	When this creature enters, it deals 3 damage to each opponent and you gain 3 life.
//	When this creature dies, discard any number of cards, then draw that many cards plus one.
var ColossusOfTheBloodAge = newColossusOfTheBloodAge

func newColossusOfTheBloodAge() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Red),
		CardFace: game.CardFace{
			Name: "Colossus of the Blood Age",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
				cost.W,
			}),
			Colors:    []color.Color{color.Red, color.White},
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Construct},
			Power:     opt.Val(game.PT{Value: 6}),
			Toughness: opt.Val(game.PT{Value: 6}),
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
								Primitive: game.Damage{
									Amount:       game.Fixed(3),
									Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(3),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentDied,
							Source:           game.TriggerSourceSelf,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.DiscardThenDraw{
									Player:     game.ControllerReference(),
									DrawOffset: 1,
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			When this creature enters, it deals 3 damage to each opponent and you gain 3 life.
			When this creature dies, discard any number of cards, then draw that many cards plus one.
		`,
		},
	}
}
