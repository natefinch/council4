package n

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// NatureSWill is the card definition for Nature's Will.
//
// Type: Enchantment
// Cost: {2}{G}{G}
//
// Oracle text:
//
//	Whenever one or more creatures you control deal combat damage to a player, tap all lands that player controls and untap all lands you control.
var NatureSWill = newNatureSWill

func newNatureSWill() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Nature's Will",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							OneOrMore:             true,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Group: game.PlayerControlledGroup(game.EventPlayerReference(), game.Selection{RequiredTypes: []types.Card{types.Land}}),
								},
							},
							{
								Primitive: game.Untap{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Whenever one or more creatures you control deal combat damage to a player, tap all lands that player controls and untap all lands you control.
		`,
		},
	}
}
