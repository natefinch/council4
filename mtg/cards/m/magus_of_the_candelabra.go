package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MagusOfTheCandelabra is the card definition for Magus of the Candelabra.
//
// Type: Creature — Human Wizard
// Cost: {G}
//
// Oracle text:
//
//	{X}, {T}: Untap X target lands.
var MagusOfTheCandelabra = newMagusOfTheCandelabra

func newMagusOfTheCandelabra() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Magus of the Candelabra",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Human, types.Wizard},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 2}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{X}, {T}: Untap X target lands.",
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
										Constraint:   "x target lands",
										Allow:        game.TargetAllowPermanent,
										Selection:    opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
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
			{X}, {T}: Untap X target lands.
		`,
		},
	}
}
