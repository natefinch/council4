package g

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// GrimDiscovery is the card definition for Grim Discovery.
//
// Type: Sorcery
// Cost: {1}{B}
//
// Oracle text:
//
//	Choose one or both —
//	• Return target creature card from your graveyard to your hand.
//	• Return target land card from your graveyard to your hand.
var GrimDiscovery = newGrimDiscovery

func newGrimDiscovery() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "Grim Discovery",
			ManaCost: opt.Val(cost.Mana{
				cost.O(1),
				cost.B,
			}),
			Colors: []color.Color{color.Black},
			Types:  []types.Card{types.Sorcery},
			SpellAbility: opt.Val(game.AbilityContent{
				Modes: []game.Mode{
					game.Mode{
						Text: "Return target creature card from your graveyard to your hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target creature card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Creature}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card: game.CardReference{
										Kind: game.CardReferenceTarget,
									},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
								},
							},
						},
					},
					game.Mode{
						Text: "Return target land card from your graveyard to your hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 1,
								MaxTargets: 1,
								Constraint: "target land card from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{RequiredTypes: []types.Card{types.Land}, Controller: game.ControllerYou}),
							},
						},
						Sequence: []game.Instruction{
							{
								Primitive: game.MoveCard{
									Card: game.CardReference{
										Kind: game.CardReferenceTarget,
									},
									FromZone:    zone.Graveyard,
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
			• Return target creature card from your graveyard to your hand.
			• Return target land card from your graveyard to your hand.
		`,
		},
	}
}
