package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SemblanceAnvil is the card definition for Semblance Anvil.
//
// Type: Artifact
// Cost: {3}
//
// Oracle text:
//
//	Imprint — When this artifact enters, you may exile a nonland card from your hand.
//	Spells you cast that share a card type with the exiled card cost {2} less to cast.
var SemblanceAnvil = newSemblanceAnvil

func newSemblanceAnvil() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Semblance Anvil",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types: []types.Card{types.Artifact},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Spells you cast that share a card type with the exiled card cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:                              game.CostModifierSpell,
								SharedExiledCardTypeReduction:     2,
								SharedExiledCardTypeReductionOnce: true,
								ExiledLinkKey:                     game.LinkedKey("exiled-with-source"),
								ExiledLinkObjectScoped:            true,
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "Imprint — When this artifact enters, you may exile a nonland card from your hand.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:           game.EventPermanentEnteredBattlefield,
							Source:          game.TriggerSourceSelf,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.ChooseFromZone{
											Player:     game.ControllerReference(),
											SourceZone: zone.Hand,
											Filter:     game.Selection{ExcludedTypes: []types.Card{types.Land}},
											Quantity:   game.Fixed(1),
											Destination: game.ChooseDestination{
												Zone: zone.Exile,
											},
											Riders: game.ChooseRiders{
												PublishLinked:       game.LinkedKey("exiled-with-source"),
												PublishObjectScoped: true,
											},
											Prompt: "Choose a card to exile",
										},
										Optional: true,
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
			Imprint — When this artifact enters, you may exile a nonland card from your hand.
			Spells you cast that share a card type with the exiled card cost {2} less to cast.
		`,
		},
	}
}
