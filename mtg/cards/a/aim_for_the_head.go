package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AimForTheHead is the card definition for Aim for the Head.
//
// Type: Sorcery
// Cost: {2}{B}
//
// Oracle text:
//
//	Choose one —
//	• Exile target Zombie.
//	• Target opponent exiles two cards from their hand.
var AimForTheHead = newAimForTheHead

func newAimForTheHead() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Aim for the Head",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Exile target Zombie.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target Zombie",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Zombie")}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MovePermanent{
									Object:      game.TargetPermanentReference(0),
									Destination: zone.Exile,
								},
							},
						},
					},
					game.Mode{
						Text: "Target opponent exiles two cards from their hand.",
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
								Primitive: game.ChooseFromZone{
									Player:     game.TargetPlayerReference(0),
									SourceZone: zone.Hand,
									Quantity:   game.Fixed(2),
									Destination: game.ChooseDestination{
										Zone: zone.Exile,
									},
									Riders: game.ChooseRiders{
										PublishObjectScoped: true,
									},
									Prompt: "Choose a card to exile",
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
			• Exile target Zombie.
			• Target opponent exiles two cards from their hand.
		`,
		},
	}
}
