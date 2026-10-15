package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// QuestForRenewal is the card definition for Quest for Renewal.
//
// Type: Enchantment
// Cost: {1}{G}
//
// Oracle text:
//
//	Whenever a creature you control becomes tapped, you may put a quest counter on this enchantment.
//	As long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step.
var QuestForRenewal = newQuestForRenewal

func newQuestForRenewal() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Quest for Renewal",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Enchantment},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "As long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step.",
					Condition: opt.Val(game.Condition{
						SourceCounterKind:      counter.Quest,
						SourceCounterKindKnown: true,
						SourceCountersAtLeast:  4,
					}),
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:               game.RuleEffectUntapDuringOtherPlayersUntapStep,
							Duration:           game.DurationPermanent,
							AffectedController: game.ControllerYou,
							PermanentTypes:     []types.Card{types.Creature},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Whenever a creature you control becomes tapped, you may put a quest counter on this enchantment.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentTapped,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
							StackObjectKind:  game.StackSpell,
							CounterKind:      counter.PlusOnePlusOne,
						},
					},
					Optional: true,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.Quest,
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
			OracleText: `
			Whenever a creature you control becomes tapped, you may put a quest counter on this enchantment.
			As long as there are four or more quest counters on this enchantment, untap all creatures you control during each other player's untap step.
		`,
		},
	}
}
