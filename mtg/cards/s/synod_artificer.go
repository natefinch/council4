package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SynodArtificer is the card definition for Synod Artificer.
//
// Type: Creature — Vedalken Artificer
// Cost: {2}{U}
//
// Oracle text:
//
//	{X}, {T}: Tap X target noncreature artifacts.
//	{X}, {T}: Untap X target noncreature artifacts.
var SynodArtificer = newSynodArtificer

func newSynodArtificer() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Synod Artificer",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.U,
			}),
			Colors:    []color.Color{color.Blue},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Vedalken, types.Artificer},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}, {T}: Tap X target noncreature artifacts.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
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
										MaxTargets:   20,
										Constraint:   "x target noncreature artifacts",
										Allow:        game.TargetAllowPermanent,
										Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Creature}}),
										CountEqualsX: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Tap{
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
				game.ActivatedAbility{
					Text: "{X}, {T}: Untap X target noncreature artifacts.",
					ManaCost: opt.Val(cost.Mana{
						cost.X,
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
										MaxTargets:   20,
										Constraint:   "x target noncreature artifacts",
										Allow:        game.TargetAllowPermanent,
										Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact}, ExcludedTypes: []types.Card{types.Creature}}),
										CountEqualsX: true,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(0),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(1),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(2),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(3),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(4),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(5),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(6),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(7),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(8),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(9),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(10),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(11),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(12),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(13),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(14),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(15),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(16),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(17),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(18),
										},
									},
									game.Instruction{
										Primitive: game.Untap{
											Object: game.TargetPermanentReference(19),
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
			{X}, {T}: Tap X target noncreature artifacts.
			{X}, {T}: Untap X target noncreature artifacts.
		`,
		},
	}
}
