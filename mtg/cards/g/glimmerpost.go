package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
)

// Glimmerpost is the card definition for Glimmerpost.
//
// Type: Land — Locus
//
// Oracle text:
//
//	When this land enters, you gain 1 life for each Locus on the battlefield.
//	{T}: Add {C}.
var Glimmerpost = newGlimmerpost

func newGlimmerpost() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Glimmerpost",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Locus},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Text: "When this land enters, you gain 1 life for each Locus on the battlefield.",
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
										Primitive: game.GainLife{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:       game.DynamicAmountCountSelector,
												Multiplier: 1,
												Group:      game.BattlefieldGroup(game.Selection{SubtypesAny: []types.Sub{types.Sub("Locus")}}),
											}),
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
			When this land enters, you gain 1 life for each Locus on the battlefield.
			{T}: Add {C}.
		`,
		},
	}
}
