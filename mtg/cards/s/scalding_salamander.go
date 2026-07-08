package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// ScaldingSalamander is the card definition for Scalding Salamander.
//
// Type: Creature — Salamander
// Cost: {2}{R}
//
// Oracle text:
//
//	Whenever this creature attacks, you may have it deal 1 damage to each creature without flying defending player controls.
var ScaldingSalamander = newScaldingSalamander

func newScaldingSalamander() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Scalding Salamander",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Salamander},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 1}),
			TriggeredAbilities: []game.TriggeredAbility{
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
									Recipient:    game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying, ControlledByDefendingPlayer: true})),
									DamageSource: opt.Val(game.EventPermanentReference()),
								},
								Optional: true,
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever this creature attacks, you may have it deal 1 damage to each creature without flying defending player controls.
		`,
		},
	}
}
