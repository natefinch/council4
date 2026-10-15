package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RoadsideReliquary is the card definition for Roadside Reliquary.
//
// Type: Land
//
// Oracle text:
//
//	{T}: Add {C}.
//	{2}, {T}, Sacrifice this land: Draw a card if you control an artifact. Draw a card if you control an enchantment.
var RoadsideReliquary = newRoadsideReliquary

func newRoadsideReliquary() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "Roadside Reliquary",
			Types: []types.Card{types.Land},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, {T}, Sacrifice this land: Draw a card if you control an artifact. Draw a card if you control an enchantment.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this land",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "if you control an artifact",
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Artifact}},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
									},
									game.Instruction{
										Primitive: game.Draw{
											Amount: game.Fixed(1),
											Player: game.ControllerReference(),
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text: "if you control an enchantment",
												ControlsMatching: opt.Val(game.SelectionCount{
													Selection: game.Selection{RequiredTypes: []types.Card{types.Enchantment}},
												}),
												SourceCounterKind: counter.PlusOnePlusOne,
											}),
										}),
									},
								},
							},
						},
						MinModes: 1,
						MaxModes: 1,
					},
				},
			},
			ManaAbilities: []game.ManaAbility{
				game.TapManaAbility(mana.C),
			},
			OracleText: `
			{T}: Add {C}.
			{2}, {T}, Sacrifice this land: Draw a card if you control an artifact. Draw a card if you control an enchantment.
		`,
		},
	}
}
