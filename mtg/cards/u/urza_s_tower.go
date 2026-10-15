package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrzaSTower is the card definition for Urza's Tower.
//
// Type: Land — Urza's Tower
//
// Oracle text:
//
//	{T}: Add {C}. If you control an Urza's Mine and an Urza's Power-Plant, add {C}{C}{C} instead.
var UrzaSTower = newUrzaSTower

func newUrzaSTower() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Urza's Tower",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Sub("Urza's"), types.Tower},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. If you control an Urza's Mine and an Urza's Power-Plant, add {C}{C}{C} instead.",
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
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                    "If you control an Urza's Mine and an Urza's Power-Plant",
												Negate:                  true,
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Power-Plant"},
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                    "If you control an Urza's Mine and an Urza's Power-Plant",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Power-Plant"},
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                    "If you control an Urza's Mine and an Urza's Power-Plant",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Power-Plant"},
											}),
										}),
									},
									game.Instruction{
										Primitive: game.AddMana{
											Amount:    game.Fixed(1),
											ManaColor: mana.C,
										},
										Condition: opt.Val(game.EffectCondition{
											Condition: opt.Val(game.Condition{
												Text:                    "If you control an Urza's Mine and an Urza's Power-Plant",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Power-Plant"},
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
			OracleText: `
			{T}: Add {C}. If you control an Urza's Mine and an Urza's Power-Plant, add {C}{C}{C} instead.
		`,
		},
	}
}
