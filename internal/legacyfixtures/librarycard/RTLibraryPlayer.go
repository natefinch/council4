// Legacy regression fixture preserved when the Oracle compiler was retired.
package fixtures

import (
	"github.com/natefinch/council4/mtg/game"
	"github.com/natefinch/council4/mtg/game/cost"
	"github.com/natefinch/council4/mtg/game/counter"
	"github.com/natefinch/council4/mtg/game/types"
	"github.com/natefinch/council4/mtg/game/zone"
	"github.com/natefinch/council4/opt"
)

// RTLibraryPlayer is the card definition for RT Library Player.
//
// Type: Artifact
//
// Oracle text:
//
//	{T}: Look at the top card of target opponent's library.
var RTLibraryPlayer = newRTLibraryPlayer

func newRTLibraryPlayer() *game.CardDef {
	return &game.CardDef{
		CardFace: game.CardFace{
			Name:  "RT Library Player",
			Types: []types.Card{types.Artifact},
			ActivatedAbilities: []game.ActivatedAbility{
				game.ActivatedAbility{
					Text: "{T}: Look at the top card of target opponent's library.",
					AdditionalCosts: []cost.Additional{
						cost.Additional{
							Kind:        cost.AdditionalTap,
							CounterKind: counter.PlusOnePlusOne,
						},
					},
					ZoneOfFunction: zone.Battlefield,
					Content: game.AbilityContent{
						Modes: []game.Mode{
							game.Mode{
								Targets: []game.TargetSpec{
									game.TargetSpec{
										MinTargets: 1,
										MaxTargets: 1,
										Constraint: "target opponent",
										Allow:      game.TargetAllowPlayer,
										Selection:  opt.Val(game.Selection{Player: game.PlayerOpponent}),
									},
								},
								Sequence: []game.Instruction{
									game.Instruction{
										Primitive: game.LookAtLibraryTop{
											Player:        game.TargetPlayerReference(0),
											PublishLinked: game.LinkedKey("look-at-target-library-top"),
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
			{T}: Look at the top card of target opponent's library.
		`,
		},
	}
}
