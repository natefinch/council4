package u

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/mana"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// UrzaSMine is the card definition for Urza's Mine.
//
// Type: Land — Urza's Mine
//
// Oracle text:
//
//	{T}: Add {C}. If you control an Urza's Power-Plant and an Urza's Tower, add {C}{C} instead.
var UrzaSMine = newUrzaSMine

func newUrzaSMine() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:     "Urza's Mine",
			Types:    []types.Card{types.Land},
			Subtypes: []types.Sub{types.Sub("Urza's"), types.Mine},
			ManaAbilities: []game.ManaAbility{
				game.ManaAbility{
					Text: "{T}: Add {C}. If you control an Urza's Power-Plant and an Urza's Tower, add {C}{C} instead.",
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
												Text:                    "If you control an Urza's Power-Plant and an Urza's Tower",
												Negate:                  true,
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Power-Plant", "Urza's Tower"},
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
												Text:                    "If you control an Urza's Power-Plant and an Urza's Tower",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Power-Plant", "Urza's Tower"},
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
												Text:                    "If you control an Urza's Power-Plant and an Urza's Tower",
												SourceCounterKind:       counter.PlusOnePlusOne,
												ControllerControlsNamed: []string{"Urza's Power-Plant", "Urza's Tower"},
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
			{T}: Add {C}. If you control an Urza's Power-Plant and an Urza's Tower, add {C}{C} instead.
		`,
		},
	}
}
