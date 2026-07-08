package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PsychicOverload is the card definition for Psychic Overload.
//
// Type: Enchantment — Aura
// Cost: {3}{U}
//
// Oracle text:
//
//	Enchant permanent
//	When this Aura enters, tap enchanted permanent.
//	Enchanted permanent doesn't untap during its controller's untap step.
//	Enchanted permanent has "Discard two artifact cards: Untap this permanent."
var PsychicOverload = newPsychicOverload

func newPsychicOverload() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Psychic Overload",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.U,
			}),
			Colors:   []color.Color{color.Blue},
			Types:    []types.Card{types.Enchantment},
			Subtypes: []types.Sub{types.Aura},
			StaticAbilities: []game.StaticAbility{
				game.EnchantStaticAbility(&game.TargetSpec{
					MinTargets: 1,
					MaxTargets: 1,
					Constraint: "permanent",
					Allow:      game.TargetAllowPermanent,
				}),
				game.StaticAbility{
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:             game.RuleEffectDoesntUntap,
							AffectedAttached: true,
						},
					},
				},
				game.StaticAbility{
					ContinuousEffects: []game.ContinuousEffect{
						game.ContinuousEffect{
							Layer: game.LayerAbility,
							Group: game.AttachedObjectGroup(game.SourcePermanentReference()),
							AddAbilities: []game.Ability{
								new(game.ActivatedAbility{
									Text: "Discard two artifact cards: Untap this permanent.",
									AdditionalCosts: []cost.Additional{
										{
											Kind:          cost.AdditionalDiscard,
											Text:          "Discard two artifact cards",
											Amount:        2,
											Source:        zone.Hand,
											MatchCardType: true,
											CardType:      types.Artifact,
										},
									},
									ZoneOfFunction: zone.Battlefield,
									Content: game.Mode{
										Sequence: []game.Instruction{
											{
												Primitive: game.Untap{
													Object: game.SourcePermanentReference(),
												},
											},
										},
									}.Ability(),
								}),
							},
						},
					},
				},
			},
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.Mode{
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.SourceAttachedPermanentReference(),
								},
							},
						},
					}.Ability(),
				},
			},
			OracleText: `
			Enchant permanent
			When this Aura enters, tap enchanted permanent.
			Enchanted permanent doesn't untap during its controller's untap step.
			Enchanted permanent has "Discard two artifact cards: Untap this permanent."
		`,
		},
	}
}
