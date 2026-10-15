package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// PharagaxGiant is the card definition for Pharagax Giant.
//
// Type: Creature — Giant
// Cost: {4}{R}
//
// Oracle text:
//
//	Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)
//	When this creature enters, if tribute wasn't paid, this creature deals 5 damage to each opponent.
var PharagaxGiant = newPharagaxGiant

func newPharagaxGiant() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Pharagax Giant",
			ManaCost: opt.Val(cost.Mana{
				cost.O(4),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Giant},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this creature enters, if tribute wasn't paid, this creature deals 5 damage to each opponent.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
						InterveningIf: "if tribute wasn't paid",
						InterveningCondition: opt.Val(game.Condition{
							Text:                 "if tribute wasn't paid",
							SourceCounterKind:    counter.PlusOnePlusOne,
							SourceTributeNotPaid: true,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Damage{
											Amount:       game.Fixed(5),
											Recipient:    game.PlayerGroupDamageRecipient(game.OpponentsReference()),
											DamageSource: opt.Val(game.SourcePermanentReference()),
										},
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.TributeReplacement("As this creature enters, an opponent of your choice may put 2 +1/+1 counters on it.", 2),
			},
			OracleText: `
			Tribute 2 (As this creature enters, an opponent of your choice may put two +1/+1 counters on it.)
			When this creature enters, if tribute wasn't paid, this creature deals 5 damage to each opponent.
		`,
		},
	}
}
