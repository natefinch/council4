package t

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// TheorixCharm is the card definition for Theorix Charm.
//
// Type: Instant
// Cost: {U}{B}
//
// Oracle text:
//
//	Choose one —
//	• Counter target noncreature spell unless its controller pays {2}.
//	• Target creature gets -2/-2 until end of turn.
//	• Mill three cards, then draw a card. (To mill three cards, put the top three cards of your library into your graveyard.)
var TheorixCharm = newTheorixCharm

func newTheorixCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue, color.Black),
		CardFace: game.CardFace{
			Name: "Theorix Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.U,
				cost.B,
			}),
			Colors: []color.Color{color.Black, color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target noncreature spell unless its controller pays {2}.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target noncreature spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									ExcludedSpellCardTypes: []types.Card{types.Creature},
									StackObjectKinds:       []game.StackObjectKind{game.StackSpell},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Pay{
									Payment: game.ResolutionPayment{
										Prompt: "Pay {2}?",
										Payer:  opt.Val(game.ObjectControllerReference(game.TargetStackObjectReference(0))),
										ManaCost: opt.Val(cost.Mana{
											cost.O(2),
										}),
									},
								},
								PublishResult: game.ResultKey("unless-paid"),
							},
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
								ResultGate: opt.Val(game.InstructionResultGate{
									Key:       "unless-paid",
									Succeeded: game.TriFalse,
								}),
							},
						},
					},
					game.Mode{
						Text: "Target creature gets -2/-2 until end of turn.",
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
								Primitive: game.ModifyPT{
									Object:         game.TargetPermanentReference(0),
									PowerDelta:     game.Fixed(-2),
									ToughnessDelta: game.Fixed(-2),
									Duration:       game.DurationUntilEndOfTurn,
								},
							},
						},
					},
					game.Mode{
						Text: "Mill three cards, then draw a card. (To mill three cards, put the top three cards of your library into your graveyard.)",
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveTopOfLibrary{
									Amount:      game.Fixed(3),
									Player:      game.ControllerReference(),
									Destination: zone.Graveyard,
								},
							},
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
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
			• Counter target noncreature spell unless its controller pays {2}.
			• Target creature gets -2/-2 until end of turn.
			• Mill three cards, then draw a card. (To mill three cards, put the top three cards of your library into your graveyard.)
		`,
		},
	}
}
