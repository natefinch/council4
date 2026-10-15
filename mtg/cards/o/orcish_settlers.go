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

// OrcishSettlers is the card definition for Orcish Settlers.
//
// Type: Creature — Orc
// Cost: {1}{R}
//
// Oracle text:
//
//	{X}{X}{R}, {T}, Sacrifice this creature: Destroy X target lands.
var OrcishSettlers = newOrcishSettlers

func newOrcishSettlers() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Orcish Settlers",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
			}),
			Colors:    []color.Color{color.Red},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Orc},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}{X}{R}, {T}, Sacrifice this creature: Destroy X target lands.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
						cost.X,
						cost.R,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
						cost.Additional{
							Kind:        cost.AdditionalSacrificeSource,
							Text:        "Sacrifice this creature",
							Amount:      1,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MaxTargets:   20,
										Constraint:   "x target lands",
										Allow:        game.TargetAllowPermanent,
										Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
										CountEqualsX: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Destroy{
											Object: game.AllTargetPermanentsReference(0),
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
			{X}{X}{R}, {T}, Sacrifice this creature: Destroy X target lands.
		`,
		},
	}
}
