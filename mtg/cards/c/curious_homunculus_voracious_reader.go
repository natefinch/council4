package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// CuriousHomunculus is the card definition for Curious Homunculus // Voracious Reader.
//
// Type: Creature — Homunculus // Creature — Eldrazi Homunculus
// Face: Voracious Reader — Creature — Eldrazi Homunculus
//
// Oracle text:
//
//	{T}: Add {C}. Spend this mana only to cast an instant or sorcery spell.
//	At the beginning of your upkeep, if there are three or more instant and/or sorcery cards in your graveyard, transform this creature.
var CuriousHomunculus = newCuriousHomunculus

func newCuriousHomunculus() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Curious Homunculus",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Homunculus},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. Spend this mana only to cast an instant or sorcery spell.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
											SpendRider: opt.Val(game.ManaSpendRider{
												Condition:   game.ManaSpendCastInstantOrSorcerySpell,
												Restriction: game.ManaSpendRestrictedToCondition,
											}),
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
					Text: "At the beginning of your upkeep, if there are three or more instant and/or sorcery cards in your graveyard, transform this creature.",
					Trigger: game.TriggerCondition{
						Type: game.TriggerAt,
						Pattern: game.TriggerPattern{
							Event:           game.EventBeginningOfStep,
							Controller:      game.TriggerControllerYou,
							StackObjectKind: game.StackSpell,
							CounterKind:     counter.PlusOnePlusOne,
							Step:            game.StepUpkeep,
						},
						InterveningIf: "if there are three or more instant and/or sorcery cards in your graveyard",
						InterveningCondition: opt.Val(game.Condition{
							Text:              "if there are three or more instant and/or sorcery cards in your graveyard",
							SourceCounterKind: counter.PlusOnePlusOne,
							ControllerGraveyardInstantOrSorceryCountAtLeast: 3,
						}),
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Transform{
											Object: game.SourcePermanentReference(),
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
			{T}: Add {C}. Spend this mana only to cast an instant or sorcery spell.
			At the beginning of your upkeep, if there are three or more instant and/or sorcery cards in your graveyard, transform this creature.
		`,
		},
		Layout: game.LayoutTransform,
		Back: opt.Val(game.CardFace{
			Name:      "Voracious Reader",
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Eldrazi, types.Homunculus},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 4}),
			StaticAbilities: []game.StaticAbility{
				game.ProwessStaticBody,
				game.StaticAbility{
					Text: "Instant and sorcery spells you cast cost {1} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Instant}},
							},
						},
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 1,
								CardSelection:    game.Selection{RequiredTypes: []types.Card{types.Sorcery}},
							},
						},
					},
				},
			},
			OracleText: `
			Prowess (Whenever you cast a noncreature spell, this creature gets +1/+1 until end of turn.)
			Instant and sorcery spells you cast cost {1} less to cast.
		`,
		}),
	}
}
