package s

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// SwirlingTorrent is the card definition for Swirling Torrent.
//
// Type: Sorcery
// Cost: {5}{U}
//
// Oracle text:
//
//	Choose one or both —
//	• Put target creature on top of its owner's library.
//	• Return target creature to its owner's hand.
var SwirlingTorrent = newSwirlingTorrent

func newSwirlingTorrent() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Blue),
		CardFace: game.CardFace{
			Name: "Swirling Torrent",
			ManaCost: opt.Val(cost.Mana{
				cost.O(5),
				cost.U,
			}),
			Colors: []color.Color{color.Blue},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Put target creature on top of its owner's library.",
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
								Primitive: game.MovePermanent{
									Object:      game.TargetPermanentReference(0),
									Destination: zone.Library,
								},
							},
						},
					},
					game.Mode{
						Text: "Return target creature to its owner's hand.",
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
								Primitive: game.MovePermanent{
									Object:      game.TargetPermanentReference(0),
									Destination: zone.Hand,
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
			• Put target creature on top of its owner's library.
			• Return target creature to its owner's hand.
		`,
		},
	}
}
