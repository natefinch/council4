package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// ArchdruidSCharm is the card definition for Archdruid's Charm.
//
// Type: Instant
// Cost: {G}{G}{G}
//
// Oracle text:
//
//	Choose one —
//	• Search your library for a creature or land card and reveal it. Put it onto the battlefield tapped if it's a land card. Otherwise, put it into your hand. Then shuffle.
//	• Put a +1/+1 counter on target creature you control. It deals damage equal to its power to target creature you don't control.
//	• Exile target artifact or enchantment.
var ArchdruidSCharm = newArchdruidSCharm

func newArchdruidSCharm() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Archdruid's Charm",
			ManaCost: opt.Val(cost.Mana{
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors: []color.Color{color.Green},
			Types:  []types.Card{types.Instant},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Search your library for a creature or land card and reveal it. Put it onto the battlefield tapped if it's a land card. Otherwise, put it into your hand. Then shuffle.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Search{
									Player: game.ControllerReference(),
									Spec: game.SearchSpec{
										SourceZone: zone.Library,
										Filter:     game.Selection{RequiredTypesAny: []types.Card{types.Creature, types.Land}},
										Reveal:     true,
										RevealOnly: true,
									},
									Amount:        game.Fixed(1),
									PublishLinked: game.LinkedKey("search-land-conditional"),
								},
							},
							{
								Primitive: game.ConditionalDestinationPlace{
									Card: game.CardReference{
										Kind:   game.CardReferenceLinked,
										LinkID: "search-land-conditional",
									},
									FromZone: zone.Library,
									CardCondition: opt.Val(game.CardSelection{
										Card: game.CardReference{
											Kind:   game.CardReferenceLinked,
											LinkID: "search-land-conditional",
										},
										Selection: game.Selection{RequiredTypes: []types.Card{types.Land}},
									}),
									EntryTapped:   true,
									ThenMandatory: true,
									Else:          zone.Hand,
								},
							},
							{
								Primitive: game.ShuffleLibrary{
									Player: game.ControllerReference(),
								},
							},
						},
					},
					game.Mode{
						Text: "Put a +1/+1 counter on target creature you control. It deals damage equal to its power to target creature you don't control.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature you don't control",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Creature}, Controller: game.ControllerNotYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.AddCounter{
									Amount:      game.Fixed(1),
									Object:      game.TargetPermanentReference(0),
									CounterKind: counter.PlusOnePlusOne,
								},
							},
							{
								Primitive: game.Damage{
									Amount: game.Dynamic(game.DynamicAmount{
										Kind:       game.DynamicAmountObjectPower,
										Multiplier: 1,
										Object:     game.TargetPermanentReference(0),
									}),
									Recipient:    game.AnyTargetDamageRecipient(1),
									DamageSource: opt.Val(game.TargetPermanentReference(0)),
								},
							},
						},
					},
					game.Mode{
						Text: "Exile target artifact or enchantment.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target artifact or enchantment",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Artifact, types.Enchantment}}),
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
				},
				MinModes: 1,
				MaxModes: 1,
			}),
			OracleText: `
			Choose one —
			• Search your library for a creature or land card and reveal it. Put it onto the battlefield tapped if it's a land card. Otherwise, put it into your hand. Then shuffle.
			• Put a +1/+1 counter on target creature you control. It deals damage equal to its power to target creature you don't control.
			• Exile target artifact or enchantment.
		`,
		},
	}
}
