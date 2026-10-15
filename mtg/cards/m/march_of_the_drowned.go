package m

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/color"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// MarchOfTheDrowned is the card definition for March of the Drowned.
//
// Type: Sorcery
// Cost: {B}
//
// Oracle text:
//
//	Choose one —
//	• Return target creature card from your graveyard to your hand.
//	• Return two target Pirate cards from your graveyard to your hand.
var MarchOfTheDrowned = newMarchOfTheDrowned

func newMarchOfTheDrowned() *game.CardDef {
	return &game.CardDef{
		ColorIdentity: color.NewIdentity(color.Black),
		CardFace: game.CardFace{
			Name: "March of the Drowned",
			ManaCost: opt.Val(cost.Mana{
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
						Text: "Return two target Pirate cards from your graveyard to your hand.",
						Targets: []game.TargetSpec{
							game.TargetSpec{
								MinTargets: 2,
								MaxTargets: 2,
								Constraint: "two target Pirate cards from your graveyard",
								Allow:      game.TargetAllowCard,
								TargetZone: zone.Graveyard,
								Selection:  opt.Val(game.Selection{SubtypesAny: []types.Sub{types.Sub("Pirate")}, Controller: game.ControllerYou}),
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
							{
								Primitive: game.MoveCard{
									Card: game.CardReference{
										Kind:        game.CardReferenceTarget,
										TargetIndex: 1,
									},
									FromZone:    zone.Graveyard,
									Destination: zone.Hand,
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
			• Return target creature card from your graveyard to your hand.
			• Return two target Pirate cards from your graveyard to your hand.
		`,
		},
	}
}
