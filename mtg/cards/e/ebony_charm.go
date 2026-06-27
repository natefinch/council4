package e

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// EbonyCharm is the card definition for Ebony Charm.
//
// Type: Instant
// Cost: {B}
//
// Oracle text:
//
//	Choose one —
//	• Target opponent loses 1 life and you gain 1 life.
//	• Exile up to three target cards from a single graveyard.
//	• Target creature gains fear until end of turn. (It can't be blocked except by artifact creatures and/or black creatures.)
var EbonyCharm = newEbonyCharm()

func newEbonyCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Ebony Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Target opponent loses 1 life and you gain 1 life.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "Target opponent",
								Allow:      game.TargetAllowPlayer,
								Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.LoseLife{
									Amount: game.Fixed(1),
									Player: game.TargetPlayerReference(0),
								},
							},
							{
								Primitive: game.GainLife{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Exile up to three target cards from a single graveyard.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets:    0,
								MaxTargets:    3,
								Constraint:    "up to three target cards from a single graveyard",
								Allow:         game.TargetAllowCard,
								TargetZone:    zone.Graveyard,
								Selection:     opt.Val(game.Selection{}),
								SameGraveyard: true,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 1},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
							{
								Primitive: game.MoveCard{
									Card:        game.CardReference{Kind: game.CardReferenceTarget, TargetIndex: 2},
									FromZone:    zone.Graveyard,
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Target creature gains fear until end of turn. (It can't be blocked except by artifact creatures and/or black creatures.)",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.ApplyContinuous{
									Object: opt.Val(game.TargetPermanentReference(0)),
									ContinuousEffects: []game.ContinuousEffect{
										game.ContinuousEffect{
											Layer: game.LayerAbility,
											AddKeywords: []game.Keyword{
												game.Fear,
											},
										},
									},
									Duration: game.DurationUntilEndOfTurn,
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
			• Target opponent loses 1 life and you gain 1 life.
			• Exile up to three target cards from a single graveyard.
			• Target creature gains fear until end of turn. (It can't be blocked except by artifact creatures and/or black creatures.)
		`,
		},
	}
}
