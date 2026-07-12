package k

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// KutzilMalametExemplar is the card definition for Kutzil, Malamet Exemplar.
//
// Type: Legendary Creature — Cat Warrior
// Cost: {1}{G}{W}
//
// Oracle text:
//
//	Your opponents can't cast spells during your turn.
//	Whenever one or more creatures you control each with power greater than its base power deals combat damage to a player, draw a card.
var KutzilMalametExemplar = newKutzilMalametExemplar

func newKutzilMalametExemplar() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.White, color.Green),
		CardFace: game.CardFace{
			Name: "Kutzil, Malamet Exemplar",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
				cost.W,
			}),
			Colors:     []color.Color{color.Green, color.White},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Cat, types.Warrior},
			Power:      opt.Val(game.PT{Value: 3}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:                           game.RuleEffectCantCastSpells,
							AffectedPlayer:                 game.PlayerOpponent,
							RestrictedDuringControllerTurn: true,
						},
					},
				},
			},
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
							DamageSourceSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, PowerAboveBase: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Your opponents can't cast spells during your turn.
			Whenever one or more creatures you control each with power greater than its base power deals combat damage to a player, draw a card.
		`,
		},
	}
}
