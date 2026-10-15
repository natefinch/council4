package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// BreathOfDarigaaz is the card definition for Breath of Darigaaz.
//
// Type: Sorcery
// Cost: {1}{R}
//
// Oracle text:
//
//	Kicker {2} (You may pay an additional {2} as you cast this spell.)
//	Breath of Darigaaz deals 1 damage to each creature without flying and each player. If this spell was kicked, it deals 4 damage to each creature without flying and each player instead.
var BreathOfDarigaaz = newBreathOfDarigaaz

func newBreathOfDarigaaz() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Breath of Darigaaz",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Kicker {2}",
					KeywordAbilities: []game.KeywordAbility{
						game.KickerKeyword{
							Cost: cost.Mana{
								cost.O(2),
							},
						},
					},
				},
			},
			SpellAbility: opt.Val(game.Mode{
				Sequence: []game.Instruction{
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
						},
						Condition: opt.Val(game.EffectCondition{
							Condition: opt.Val(game.Condition{
								Negate:         true,
								SpellWasKicked: true,
							}),
						}),
						PublishCondition: "condition-0",
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(1),
							Recipient: game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
						},
						ConditionGate: "condition-0",
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, ExcludedKeyword: game.Flying})),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
					{
						Primitive: game.Damage{
							Amount:    game.Fixed(4),
							Recipient: game.PlayerGroupDamageRecipient(game.AllPlayersReference()),
						},
						ConditionGate:       "condition-0",
						ConditionGateNegate: true,
					},
				},
			}.Ability()),
			OracleText: `
			Kicker {2} (You may pay an additional {2} as you cast this spell.)
			Breath of Darigaaz deals 1 damage to each creature without flying and each player. If this spell was kicked, it deals 4 damage to each creature without flying and each player instead.
		`,
		},
	}
}
