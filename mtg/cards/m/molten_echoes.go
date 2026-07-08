package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// MoltenEchoes is the card definition for Molten Echoes.
//
// Type: Enchantment
// Cost: {2}{R}{R}
//
// Oracle text:
//
//	As this enchantment enters, choose a creature type.
//	Whenever a nontoken creature you control of the chosen type enters, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.
var MoltenEchoes = newMoltenEchoes

func newMoltenEchoes() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Molten Echoes",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Enchantment},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:            game.EventPermanentEnteredBattlefield,
							Controller:       game.TriggerControllerYou,
							SubjectSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}, SubtypeChoice: game.SubtypeChoiceSourceEntry, NonToken: true},
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.CreateToken{
									Amount: game.Fixed(1),
									Source: game.TokenCopyOf(game.TokenCopySpec{
										Source:      game.TokenCopySourceObject,
										Object:      game.EventPermanentReference(),
										AddKeywords: []game.Keyword{game.Haste},
									}),
								},
							},
							{
								Primitive: game.CreateDelayedTrigger{
									Trigger: game.DelayedTriggerDef{
										Timing: game.DelayedAtBeginningOfNextEndStep,
										Content: game.Mode{
											Sequence: []game.Instruction{
												{
													Primitive: game.Exile{
														Object: game.EventPermanentReference(),
													},
												},
											},
										}.Ability(),
									},
								},
							},
						},
					}.Ability(),
				},
			},
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntryTypeChoiceReplacement("As this enchantment enters, choose a creature type."),
			},
			OracleText: `
			As this enchantment enters, choose a creature type.
			Whenever a nontoken creature you control of the chosen type enters, create a token that's a copy of that creature. That token gains haste. Exile it at the beginning of the next end step.
		`,
		},
	}
}
