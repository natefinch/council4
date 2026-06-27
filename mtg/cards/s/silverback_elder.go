package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SilverbackElder is the card definition for Silverback Elder.
//
// Type: Creature — Ape Shaman
// Cost: {2}{G}{G}{G}
//
// Oracle text:
//
//	Whenever you cast a creature spell, choose one —
//	• Destroy target artifact or enchantment.
//	• Look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.
//	• You gain 4 life.
var SilverbackElder = newSilverbackElder()

func newSilverbackElder() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Green),
		CardFace: game.CardFace{
			Name: "Silverback Elder",
			ManaCost: opt.Val(cost.Mana{
				cost.O(2),
				cost.G,
				cost.G,
				cost.G,
			}),
			Colors:    []color.Color{color.Green},
			Types:     []types.Card{types.Creature},
			Subtypes:  []types.Sub{types.Ape, types.Shaman},
			Power:     opt.Val(game.PT{Value: 5}),
			Toughness: opt.Val(game.PT{Value: 7}),
			TriggeredAbilities: []game.TriggeredAbility{
				game.TriggeredAbility{
					Trigger: game.TriggerCondition{
						Type: game.TriggerWhenever,
						Pattern: game.TriggerPattern{
							Event:         game.EventSpellCast,
							Controller:    game.TriggerControllerYou,
							CardSelection: game.Selection{RequiredTypes: []types.Card{types.Creature}},
						},
					},
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Text: "Destroy target artifact or enchantment.",
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
										Primitive: game.Destroy{
											Object: game.TargetPermanentReference(0),
										},
									},
								},
							},
							game.Mode{
								Text: "Look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.",
								Sequence: []game.Instruction{
									{
										Primitive: game.Dig{
											Player:       game.ControllerReference(),
											Look:         game.Fixed(5),
											Take:         game.Fixed(1),
											Remainder:    game.DigRemainderLibraryBottom,
											Filter:       opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}}),
											TakeUpTo:     true,
											Destination:  zone.Battlefield,
											EntersTapped: true,
										},
									},
								},
							},
							game.Mode{
								Text: "You gain 4 life.",
								Sequence: []game.Instruction{
									{
										Primitive: game.GainLife{
											Amount: game.Fixed(4),
											Player: game.ControllerReference(),
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
			Whenever you cast a creature spell, choose one —
			• Destroy target artifact or enchantment.
			• Look at the top five cards of your library. You may put a land card from among them onto the battlefield tapped. Put the rest on the bottom of your library in a random order.
			• You gain 4 life.
		`,
		},
	}
}
