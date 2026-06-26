package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BragoKingEternal is the card definition for Brago, King Eternal.
//
// Type: Legendary Creature — Spirit Noble
// Cost: {2}{W}{U}
//
// Oracle text:
//
//	Flying
//	Whenever Brago deals combat damage to a player, exile any number of target nonland permanents you control, then return those cards to the battlefield under their owner's control.
var BragoKingEternal = newBragoKingEternal()

func newBragoKingEternal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Blue),
		CardFace: game.CardFace{
			Name: "Brago, King Eternal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.W,
				cost.U,
			}),
			Colors:     []color.Color{color.Blue, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Spirit, types.Noble},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
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
					Content: game.Mode{
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 0,
								MaxTargets: 99,
								Constraint: "any number of target nonland permanents you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Exile{
									Object:         game.AllTargetPermanentsReference(0),
									ExileLinkedKey: game.LinkedKey("group-blink"),
								},
							},
							{
								Primitive: game.PutOnBattlefield{
									Source: game.LinkedBattlefieldSource(game.LinkedKey("group-blink")),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Flying
			Whenever Brago deals combat damage to a player, exile any number of target nonland permanents you control, then return those cards to the battlefield under their owner's control.
		`,
		},
	}
}
