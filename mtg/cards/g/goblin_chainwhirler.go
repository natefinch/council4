package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// GoblinChainwhirler is the card definition for Goblin Chainwhirler.
//
// Type: Creature — Goblin Warrior
// Cost: {R}{R}{R}
//
// Oracle text:
//
//	First strike
//	When this creature enters, it deals 1 damage to each opponent and each creature and planeswalker they control.
var GoblinChainwhirler = newGoblinChainwhirler

func newGoblinChainwhirler() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Goblin Chainwhirler",
			ManaCost: opt.Val(cost.Mana{
				cost.R,
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Goblin, types.Warrior},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.FirstStrikeStaticBody,
			},
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
									Amount:       game.Fixed(1),
									Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Planeswalker}, Controller: game.ControllerOpponent})),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			First strike
			When this creature enters, it deals 1 damage to each opponent and each creature and planeswalker they control.
		`,
		},
	}
}
