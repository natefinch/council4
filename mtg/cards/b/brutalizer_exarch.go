package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BrutalizerExarch is the card definition for Brutalizer Exarch.
//
// Type: Creature — Phyrexian Cleric
// Cost: {5}{G}
//
// Oracle text:
//
//	When this creature enters, choose one —
//	• Search your library for a creature card, reveal it, then shuffle and put that card on top.
//	• Put target noncreature permanent on the bottom of its owner's library.
var BrutalizerExarch = newBrutalizerExarch()

func newBrutalizerExarch() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Brutalizer Exarch",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Phyrexian, types.Cleric},
			Power:     opt.Val(game.PT{Value: 3}),
			Toughness: opt.Val(game.PT{Value: 3}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhen,
						Pattern: game.TriggerPattern{
							Event:  game.EventPermanentEnteredBattlefield,
							Source: game.TriggerSourceSelf,
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Search your library for a creature card, reveal it, then shuffle and put that card on top.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Search{
											Player: game.ControllerReference(),
											Spec: game.SearchSpec{
												SourceZone:          zone.Library,
												Destination:         zone.Library,
												DestinationPosition: game.SearchPositionTop,
												Filter:              game.Selection{RequiredTypes: []types.Card{types.Creature}},
												Reveal:              true,
											},
											Amount: game.Fixed(1),
										},
									},
								},
							},
							game.Mode{
								Text: "Put target noncreature permanent on the bottom of its owner's library.",
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target noncreature permanent",
										Allow:      game.TargetAllowPermanent,
										Selection:  opt.Val(game.Selection{ExcludedTypes: []types.Card{types.Creature}}),
									},
								},
								Sequence: []game.Instruction{
									{
										Primitive: game.PutPermanentOnLibrary{
											Object: game.TargetPermanentReference(0),
											Bottom: true,
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
			When this creature enters, choose one —
			• Search your library for a creature card, reveal it, then shuffle and put that card on top.
			• Put target noncreature permanent on the bottom of its owner's library.
		`,
		},
	}
}
