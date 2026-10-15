package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EyeOfUgin is the card definition for Eye of Ugin.
//
// Type: Legendary Land
//
// Oracle text:
//
//	Colorless Eldrazi spells you cast cost {2} less to cast.
//	{7}, {T}: Search your library for a colorless creature card, reveal it, put it into your hand, then shuffle.
var EyeOfUgin = newEyeOfUgin

func newEyeOfUgin() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:       "Eye of Ugin",
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Land},
			StaticAbilities: []game.StaticAbility{
				game.StaticAbility{
					Text: "Colorless Eldrazi spells you cast cost {2} less to cast.",
					RuleEffects: []game.RuleEffect{
						game.RuleEffect{
							Kind:           game.RuleEffectCostModifier,
							Duration:       game.DurationPermanent,
							AffectedPlayer: game.PlayerYou,
							CostModifier: game.CostModifier{
								Kind:             game.CostModifierSpell,
								GenericReduction: 2,
								CardSelection:    game.Selection{SubtypesAny: []types.Sub{types.Eldrazi}, Colorless: true},
							},
						},
					},
				},
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{7}, {T}: Search your library for a colorless creature card, reveal it, put it into your hand, then shuffle.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(7),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:  zone.Library,
												Destination: zone.Hand,
												Filter:      game.Selection{RequiredTypes: []types.Card{types.Creature}, Colorless: true},
												Reveal:      true,
											},
											Amount: game.Fixed(1),
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
			Colorless Eldrazi spells you cast cost {2} less to cast.
			{7}, {T}: Search your library for a colorless creature card, reveal it, put it into your hand, then shuffle.
		`,
		},
	}
}
