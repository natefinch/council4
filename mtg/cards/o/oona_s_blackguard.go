package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// OonaSBlackguard is the card definition for Oona's Blackguard.
//
// Type: Creature — Faerie Rogue
// Cost: {1}{B}
//
// Oracle text:
//
//	Flying
//	Each other Rogue creature you control enters with an additional +1/+1 counter on it.
//	Whenever a creature you control with a +1/+1 counter on it deals combat damage to a player, that player discards a card.
var OonaSBlackguard = newOonaSBlackguard

func newOonaSBlackguard() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Oona's Blackguard",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Faerie, types.Rogue},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:                 game.EventDamageDealt,
							Controller:            game.TriggerControllerYou,
							Subject:               game.TriggerSubjectDamageSource,
							RequireCombatDamage:   true,
							DamageRecipient:       game.DamageRecipientPlayer,
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, MatchCounter: true, RequiredCounter: counter.PlusOnePlusOne},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Discard{
									Amount: game.Fixed(1),
									Player: game.EventPlayerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersGroupReplacement("Each other Rogue creature you control enters with an additional +1/+1 counter on it.", &game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypesAny: []types.Sub{types.Sub("Rogue")}, Controller: game.ControllerYou, ExcludeSource: true}, game.CounterPlacement{Kind: counter.PlusOnePlusOne, Amount: 1}),
			},
			OracleText: `
			Flying
			Each other Rogue creature you control enters with an additional +1/+1 counter on it.
			Whenever a creature you control with a +1/+1 counter on it deals combat damage to a player, that player discards a card.
		`,
		},
	}
}
