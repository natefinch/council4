package q

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// QuinaQuGourmet is the card definition for Quina, Qu Gourmet.
//
// Type: Legendary Creature — Qu
// Cost: {2}{G}
//
// Oracle text:
//
//	If one or more tokens would be created under your control, those tokens plus a 1/1 green Frog creature token are created instead.
//	{2}, Sacrifice a Frog: Put a +1/+1 counter on Quina.
var QuinaQuGourmet = newQuinaQuGourmet

func newQuinaQuGourmet() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Quina, Qu Gourmet",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
			}),
			Colors:     []color.Color{color.Green},
			Supertypes: []types.Super{types.Legendary},
			Types:      []types.Card{types.Creature},
			Subtypes:   []types.Sub{types.Qu},
			Power:      opt.Val(game.PT{Value: 2}),
			Toughness:  opt.Val(game.PT{Value: 3}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}, Sacrifice a Frog: Put a +1/+1 counter on Quina.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalSacrifice,
							Text:        "Sacrifice a Frog",
							Amount:      1,
							SubtypesAny: cost.SubtypeSet{0: types.Frog},
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.AddCounter{
											Amount:      game.Fixed(1),
											Object:      game.SourcePermanentReference(),
											CounterKind: counter.PlusOnePlusOne,
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
				game.TokenCreationReplacementFiltered("If one or more tokens would be created under your control, those tokens plus a 1/1 green Frog creature token are created instead.", &game.TokenCreationReplacementSpec{Multiplier: 1, Addend: 1, AddendDef: quinaQuGourmetToken, Filter: game.TriggerControllerYou}),
			},
			OracleText: `
			If one or more tokens would be created under your control, those tokens plus a 1/1 green Frog creature token are created instead.
			{2}, Sacrifice a Frog: Put a +1/+1 counter on Quina.
		`,
		},
	}
}

var quinaQuGourmetToken = newQuinaQuGourmetToken()

func newQuinaQuGourmetToken() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:      "Frog",
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Frog},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
		},
	}
}
