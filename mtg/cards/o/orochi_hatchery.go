package o

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// OrochiHatchery is the card definition for Orochi Hatchery.
//
// Type: Artifact
// Cost: {X}{X}
//
// Oracle text:
//
//	This artifact enters with X charge counters on it.
//	{5}, {T}: Create a 1/1 green Snake creature token for each charge counter on this artifact.
var OrochiHatchery = newOrochiHatchery

func newOrochiHatchery() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Orochi Hatchery",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.X,
			}),
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{5}, {T}: Create a 1/1 green Snake creature token for each charge counter on this artifact.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
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
										Primitive: game.CreateToken{
											Amount: game.Dynamic(game.DynamicAmount{
												Kind:        game.DynamicAmountObjectCounters,
												Multiplier:  1,
												CounterKind: counter.Charge,
												Object:      game.SourcePermanentReference(),
											}),
											Source: game.TokenDef(orochiHatcheryToken),
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
			ReplacementAbilities: []game.ReplacementAbility{
				game.EntersWithCountersReplacement("This artifact enters with X charge counters on it.", game.CounterPlacement{Kind: counter.Charge, AmountFromX: true}),
			},
			OracleText: `
			This artifact enters with X charge counters on it.
			{5}, {T}: Create a 1/1 green Snake creature token for each charge counter on this artifact.
		`,
		},
	}
}

var orochiHatcheryToken = newOrochiHatcheryToken()

func newOrochiHatcheryToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Snake",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Snake},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
