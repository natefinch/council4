package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GhostLitStalker is the card definition for Ghost-Lit Stalker.
//
// Type: Creature — Spirit
// Cost: {B}
//
// Oracle text:
//
//	{4}{B}, {T}: Target player discards two cards. Activate only as a sorcery.
//	Channel — {5}{B}{B}, Discard this card: Target player discards four cards. Activate only as a sorcery.
var GhostLitStalker = newGhostLitStalker

func newGhostLitStalker() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ghost-Lit Stalker",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors:    []color.Color{color.Black},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Spirit},
			Power:     opt.Val(game.PT{Value: 1}),
			Toughness: opt.Val(game.PT{Value: 1}),
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{4}{B}, {T}: Target player discards two cards. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(4),
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(2),
											Player: game.TargetPlayerReference(0),
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
					Text: "Channel — {5}{B}{B}, Discard this card: Target player discards four cards. Activate only as a sorcery.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(5),
						cost.B,
						cost.B,
					}),
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalDiscard,
							Text:        "Discard this card",
							Amount:      1,
							Source:      zone.Hand,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Hand,
					Timing:         game.SorceryOnly,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "Target player",
										Allow:      game.TargetAllowPlayer,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.Discard{
											Amount: game.Fixed(4),
											Player: game.TargetPlayerReference(0),
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
			{4}{B}, {T}: Target player discards two cards. Activate only as a sorcery.
			Channel — {5}{B}{B}, Discard this card: Target player discards four cards. Activate only as a sorcery.
		`,
		},
	}
}
