package c

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// CrypticCommand is the card definition for Cryptic Command.
//
// Type: Instant
// Cost: {1}{U}{U}{U}
//
// Oracle text:
//
//	Choose two —
//	• Counter target spell.
//	• Return target permanent to its owner's hand.
//	• Tap all creatures your opponents control.
//	• Draw a card.
var CrypticCommand = newCrypticCommand

func newCrypticCommand() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Cryptic Command",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.U,
				cost.U,
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Counter target spell.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target spell",
								Allow:      game.TargetAllowStackObject,
								Predicate: game.TargetPredicate{
									StackObjectKinds: []game.StackObjectKind{game.StackSpell},
								},
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.CounterObject{
									Object: game.TargetStackObjectReference(0),
								},
							},
						},
					},
					game.Mode{
						Text: "Return target permanent to its owner's hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target permanent",
								Allow:      game.TargetAllowPermanent,
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Object:      game.TargetPermanentReference(0),
									Destination: zone.Hand,
								},
							},
						},
					},
					game.Mode{
						Text: "Tap all creatures your opponents control.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Tap{
									Group: game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerOpponent}),
								},
							},
						},
					},
					game.Mode{
						Text: "Draw a card.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Draw{
									Amount: game.Fixed(1),
									Player: game.ControllerReference(),
								},
							},
						},
					},
				},
				MinModes: 2,
				MaxModes: 2,
			}),
			OracleText: `
			Choose two —
			• Counter target spell.
			• Return target permanent to its owner's hand.
			• Tap all creatures your opponents control.
			• Draw a card.
		`,
		},
	}
}
