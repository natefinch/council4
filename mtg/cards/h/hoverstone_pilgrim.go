package h

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// HoverstonePilgrim is the card definition for Hoverstone Pilgrim.
//
// Type: Artifact Creature — Golem
// Cost: {5}
//
// Oracle text:
//
//	Flying
//	Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
//	{2}: Put target card from a graveyard on the bottom of its owner's library.
var HoverstonePilgrim = newHoverstonePilgrim

func newHoverstonePilgrim() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Hoverstone Pilgrim",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Golem},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 5}),
			StaticAbilities: []game.StaticAbility{
				game.FlyingStaticBody,
				game.WardStaticAbility(cost.Mana{cost.O(2)}),
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: Put target card from a graveyard on the bottom of its owner's library.",
					ManaCost: opt.Val(cost.Mana{
						cost.O(2),
					}),
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target card from a graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{}),
										TargetZone: zone.Graveyard,
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.MoveCard{
											Card: game.CardReference{
												Kind: game.CardReferenceTarget,
											},
											FromZone:          zone.Graveyard,
											Destination:       zone.Library,
											DestinationBottom: true,
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
			Flying
			Ward {2} (Whenever this creature becomes the target of a spell or ability an opponent controls, counter it unless that player pays {2}.)
			{2}: Put target card from a graveyard on the bottom of its owner's library.
		`,
		},
	}
}
