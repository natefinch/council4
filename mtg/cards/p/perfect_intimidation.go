package p

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// PerfectIntimidation is the card definition for Perfect Intimidation.
//
// Type: Sorcery
// Cost: {3}{B}
//
// Oracle text:
//
//	Choose one or both —
//	• Target opponent exiles two cards from their hand.
//	• Remove all counters from target creature.
var PerfectIntimidation = newPerfectIntimidation()

func newPerfectIntimidation() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Perfect Intimidation",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
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
									Filter:     game.Selection{},
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
					game.Mode{
						Text: "Remove all counters from target creature.",
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
								Primitive: game.RemoveCounter{
									Object:   game.TargetPermanentReference(0),
									AllKinds: true,
								},
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Choose one or both —
			• Target opponent exiles two cards from their hand.
			• Remove all counters from target creature.
		`,
		},
	}
}
