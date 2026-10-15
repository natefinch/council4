package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheEchoverseFulcrum is the card definition for The Echoverse Fulcrum.
//
// Type: Legendary Artifact
// Cost: {2}
//
// Oracle text:
//
//	When The Echoverse Fulcrum enters, draw a card, then discard a card.
//	{5}, {T}, Exile The Echoverse Fulcrum: Destroy all creatures. Activate only as a sorcery.
var TheEchoverseFulcrum = newTheEchoverseFulcrum

func newTheEchoverseFulcrum() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "The Echoverse Fulcrum",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
			}),
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, {T}, Exile The Echoverse Fulcrum: Destroy all creatures. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalExileSource,
							Text:        "Exile The Echoverse Fulcrum",
							Amount:      1,
							Source:      zone.Battlefield,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}}),
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
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When The Echoverse Fulcrum enters, draw a card, then discard a card.",
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
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
									},
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
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
			When The Echoverse Fulcrum enters, draw a card, then discard a card.
			{5}, {T}, Exile The Echoverse Fulcrum: Destroy all creatures. Activate only as a sorcery.
		`,
		},
	}
}
