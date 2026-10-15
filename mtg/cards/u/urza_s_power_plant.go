package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrzaSPowerPlant is the card definition for Urza's Power Plant.
//
// Type: Land — Urza's Power-Plant
//
// Oracle text:
//
//	{T}: Add {C}. If you control an Urza's Mine and an Urza's Tower, add {C}{C} instead.
var UrzaSPowerPlant = newUrzaSPowerPlant

func newUrzaSPowerPlant() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Urza's Power Plant",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Sub("Urza's"), types.PowerPlant},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. If you control an Urza's Mine and an Urza's Tower, add {C}{C} instead.",
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
												Text:                    "If you control an Urza's Mine and an Urza's Tower",
												Negate:                  true,
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Tower"},
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
												Text:                    "If you control an Urza's Mine and an Urza's Tower",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Tower"},
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
												Text:                    "If you control an Urza's Mine and an Urza's Tower",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Mine", "Urza's Tower"},
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
			{T}: Add {C}. If you control an Urza's Mine and an Urza's Tower, add {C}{C} instead.
		`,
		},
	}
}
