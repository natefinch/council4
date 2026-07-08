package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RoninCliffrider is the card definition for Ronin Cliffrider.
//
// Type: Creature — Human Samurai
// Cost: {3}{R}{R}
//
// Oracle text:
//
//	Bushido 1 (Whenever this creature blocks or becomes blocked, it gets +1/+1 until end of turn.)
//	Whenever this creature attacks, you may have it deal 1 damage to each creature defending player controls.
var RoninCliffrider = newRoninCliffrider

func newRoninCliffrider() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Ronin Cliffrider",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.R,
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Samurai},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 2}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:      game.EventBlockerDeclared,
							Source:     game.TriggerSourceSelf,
							UnionEvent: game.EventAttackerBecameBlocked,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.ModifyPT{
									Object:         game.EventPermanentReference(),
									PowerDelta:     game.Fixed(1),
									ToughnessDelta: game.Fixed(1),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					}.Ability(),
				},
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:  game.EventAttackerDeclared,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:       game.Fixed(1),
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ControlledByDefendingPlayer: true})),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Bushido 1 (Whenever this creature blocks or becomes blocked, it gets +1/+1 until end of turn.)
			Whenever this creature attacks, you may have it deal 1 damage to each creature defending player controls.
		`,
		},
	}
}
