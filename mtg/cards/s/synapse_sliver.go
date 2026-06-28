package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// SynapseSliver is the card definition for Synapse Sliver.
//
// Type: Creature — Sliver
// Cost: {4}{U}
//
// Oracle text:
//
//	Whenever a Sliver deals combat damage to a player, its controller may draw a card.
var SynapseSliver = newSynapseSliver()

func newSynapseSliver() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Synapse Sliver",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Sliver},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{SubtypesAny: []types.Sub{types.Sub("Sliver")}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ObjectControllerReference(game.EventPermanentReference()),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ObjectControllerReference(game.EventPermanentReference())),
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever a Sliver deals combat damage to a player, its controller may draw a card.
		`,
		},
	}
}
