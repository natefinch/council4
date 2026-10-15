package r

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/opt"
)

// RealitySpasm is the card definition for Reality Spasm.
//
// Type: Instant
// Cost: {X}{U}{U}
//
// Oracle text:
//
//	Choose one —
//	• Tap X target permanents.
//	• Untap X target permanents.
var RealitySpasm = newRealitySpasm

func newRealitySpasm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Reality Spasm",
			ManaCost: opt.Val(cost.Mana{
				cost.X,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Tap X target permanents.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets:   0,
								MaxTargets:   20,
								Constraint:   "x target permanents",
								Allow:        game.TargetAllowPermanent,
								CountEqualsX: true,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Object: game.AllTargetPermanentsReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Untap X target permanents.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets:   0,
								MaxTargets:   20,
								Constraint:   "x target permanents",
								Allow:        game.TargetAllowPermanent,
								CountEqualsX: true,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(1),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(2),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(3),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(4),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(5),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(6),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(7),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(8),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(9),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(10),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(11),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(12),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(13),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(14),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(15),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(16),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(17),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(18),
								},
							},
							{
								Primitive: game.Untap{
									Object: game.TargetPermanentReference(19),
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Tap X target permanents.
			• Untap X target permanents.
		`,
		},
	}
}
