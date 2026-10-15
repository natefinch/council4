package b

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// BarkformHarvester is the card definition for Barkform Harvester.
//
// Type: Artifact Creature — Shapeshifter
// Cost: {3}
//
// Oracle text:
//
//	Changeling (This card is every creature type.)
//	Reach
//	{2}: Put target card from your graveyard on the bottom of your library.
var BarkformHarvester = newBarkformHarvester

func newBarkformHarvester() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name: "Barkform Harvester",
			ManaCost: opt.Val(cost.Mana{
				cost.O(3),
			}),
			Types:     []types.Card{types.Artifact, types.Creature},
			Subtypes:  []types.Sub{types.Shapeshifter},
			Power:     opt.Val(game.PT{Value: 2}),
			Toughness: opt.Val(game.PT{Value: 3}),
			StaticAbilities: []game.StaticAbility{
				game.ChangelingStaticBody,
				game.ReachStaticBody,
			},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{2}: Put target card from your graveyard on the bottom of your library.",
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
										Constraint: "target card from your graveyard",
										Allow:      game.TargetAllowCard,
										Selection:  opt.Val(game.Selection{Controller: game.ControllerYou}),
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
			Changeling (This card is every creature type.)
			Reach
			{2}: Put target card from your graveyard on the bottom of your library.
		`,
		},
	}
}
