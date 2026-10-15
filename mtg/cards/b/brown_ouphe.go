package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BrownOuphe is the card definition for Brown Ouphe.
//
// Type: Creature — Ouphe
// Cost: {G}
//
// Oracle text:
//
//	{1}{G}, {T}: Counter target activated ability from an artifact source. (Mana abilities can't be targeted.)
var BrownOuphe = newBrownOuphe

func newBrownOuphe() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Brown Ouphe",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ouphe},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{1}{G}, {T}: Counter target activated ability from an artifact source. (Mana abilities can't be targeted.)",
					ManaCost: opt.Val(cost.Mana{
						cost.O(1),
						cost.G,
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
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target activated ability from an artifact source",
										Allow:      game.TargetAllowStackObject,
										Predicate: game.TargetPredicate{
											StackObjectKinds:       []game.StackObjectKind{game.StackActivatedAbility},
											StackObjectSourceTypes: []types.Card{types.Artifact},
										},
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.CounterObject{
											Object: game.TargetStackObjectReference(0),
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
			{1}{G}, {T}: Counter target activated ability from an artifact source. (Mana abilities can't be targeted.)
		`,
		},
	}
}
