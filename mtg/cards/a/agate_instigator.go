package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// AgateInstigator is the card definition for AgateInstigator.
//
// Type: Creature — Lizard Rogue
// Cost: {1}{R}
//
// Oracle text:
//
//	Offspring {1}{R} (You may pay an additional {1}{R} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
//	Whenever another creature you control enters, this creature deals 1 damage to each opponent.
var AgateInstigator = newAgateInstigator

func newAgateInstigator() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Agate Instigator",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Lizard, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.OffspringStaticAbility(cost.Mana{cost.O(1), cost.R}),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.OffspringEnterTriggeredAbility(),
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							ExcludeSelf:      true,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
									DamageSource: opt.Val(game.SourcePermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Offspring {1}{R} (You may pay an additional {1}{R} as you cast this spell. If you do, when this creature enters, create a 1/1 token copy of it.)
			Whenever another creature you control enters, this creature deals 1 damage to each opponent.
		`,
		},
	}
}
