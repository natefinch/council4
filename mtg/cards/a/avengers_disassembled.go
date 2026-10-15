package a

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// AvengersDisassembled is the card definition for Avengers Disassembled.
//
// Type: Sorcery
// Cost: {1}{R}{R}
//
// Oracle text:
//
//	Choose one or both —
//	• Avengers Disassembled deals 3 damage to each creature.
//	• Destroy target land. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.
var AvengersDisassembled = newAvengersDisassembled

func newAvengersDisassembled() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Red),
		CardFace: game.CardFace{
			Name: "Avengers Disassembled",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.R,
				cost.R,
			}),
			Colors: []color.Color{color.Red},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Avengers Disassembled deals 3 damage to each creature.",
						Sequence: []game.Instruction{
							{
								Primitive: game.Damage{
									Amount:    game.Fixed(3),
									Recipient: game.GroupDamageRecipient(game.BattlefieldGroup(game.Selection{RequiredTypes: []types.Card{types.Creature}})),
								},
							},
						},
					},
					game.Mode{
						Text: "Destroy target land. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land",
								Allow:      game.TargetAllowPermanent,
								Selection:  opt.Val(game.Selection{RequiredTypesAny: []types.Card{types.Land}}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.Destroy{
									Object: game.TargetPermanentReference(0),
								},
							},
							{
								Primitive: game.Search{
									Player: game.ObjectControllerReference(game.TargetPermanentReference(0)),
									Spec: game.SearchSpec{
										SourceZone:   zone.Library,
										Destination:  zone.Battlefield,
										Filter:       game.Selection{RequiredTypes: []types.Card{types.Land}, Supertypes: []types.Super{types.Basic}},
										EntersTapped: true,
									},
									Amount: game.Fixed(1),
								},
								Optional:      true,
								OptionalActor: opt.Val(game.ObjectControllerReference(game.TargetPermanentReference(0))),
							},
						},
					},
				},
				MinModes: 1,
				MaxModes: 2,
			}),
			OracleText: `
			Choose one or both —
			• Avengers Disassembled deals 3 damage to each creature.
			• Destroy target land. Its controller may search their library for a basic land card, put it onto the battlefield tapped, then shuffle.
		`,
		},
	}
}
